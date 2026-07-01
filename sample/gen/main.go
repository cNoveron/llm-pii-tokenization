package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// WillData holds the dynamic information to inject into the Will template.
type WillData struct {
	TestatorName            string
	City                    string
	State                   string
	SpouseName              string
	Children                []string
	AlternateExecutor       string
	RealProperty            string
	CoOwner                 string
	MortgageAmount          string
	LivingTrustName         string
	LivingTrustDate         string
	SpecialNeedsBeneficiary string
	ContingencyAge          string
	Bequests                []struct {
		Amount      string
		Beneficiary string
	}
	EstateValue string
	Date        string
}

// poaLines remains the same hardcoded string for the Power of Attorney sample
var poaLines = []string{
	"DURABLE POWER OF ATTORNEY",
	"Dated August 2, 2021",
	"",
	"I, Margaret Whitfield, residing at 91 Cedar Street, Oakland, California,",
	"hereby appoint David Whitfield as my Attorney-in-Fact (Agent).",
	"Social Security Number: 987-65-4321",
	"Date of Birth: 11/03/1948",
	"Contact: margaret.whitfield@example.com, phone (628) 555-0142",
	"",
	"POWERS GRANTED",
	"My Agent, David Whitfield, is authorized to manage my financial affairs,",
	"including banking, real estate transactions, and the payment of debts up",
	"to $75,000.00 without further authorization.",
	"",
	"LIMITATIONS",
	"This Power of Attorney is durable and shall not be affected by my",
	"subsequent disability or incapacity. It takes effect immediately.",
	"My attorney of record is Sarah Chen, who prepared this instrument.",
	"",
	"IN WITNESS WHEREOF, I have executed this Durable Power of Attorney.",
	"Signed: Margaret Whitfield",
}

func main() {
	outDir := "sample"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	// 1. Generate will.pdf using HTML templates
	generateWillPDF(outDir)

	// 2. Generate poa.pdf using the original raw PDF method
	poaPath := filepath.Join(outDir, "poa.pdf")
	if err := os.WriteFile(poaPath, buildRawPDF(poaLines), 0o644); err != nil {
		log.Fatalf("write %s: %v", poaPath, err)
	}
	log.Printf("wrote %s", poaPath)
}

func generateWillPDF(outDir string) {
	// Our mock data for the Will
	data := WillData{
		TestatorName:            "John Smith",
		City:                    "San Francisco",
		State:                   "California",
		SpouseName:              "Margaret Smith",
		Children:                []string{"Emily Smith", "Thomas Smith"},
		AlternateExecutor:       "Robert Chen",
		RealProperty:            "123 Maple Street, San Francisco, CA",
		CoOwner:                 "Margaret Smith",
		MortgageAmount:          "$150,000",
		LivingTrustName:         "The John and Margaret Smith Revocable Living Trust",
		LivingTrustDate:         "January 15, 2015",
		SpecialNeedsBeneficiary: "Thomas Smith",
		ContingencyAge:          "25",
		Bequests: []struct {
			Amount      string
			Beneficiary string
		}{
			{"$500,000.00", "Emily Smith"},
			{"$250,000.00", "Thomas Smith"},
		},
		EstateValue: "$1,200,000",
		Date:        "March 14, 2019",
	}

	// Parse the HTML template we created
	tmpl, err := template.ParseFiles("sample/gen/will.html")
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	// Execute the template with our data
	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	// Create headless Chrome context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuffer []byte
	// Render the HTML into PDF
	err = chromedp.Run(ctx,
		// Load the HTML content directly
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlBuf.String()).Do(ctx)
		}),
		// Wait for body to be ready, then print to PDF
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)
	if err != nil {
		log.Fatalf("failed to generate PDF with chromedp: %v", err)
	}

	path := filepath.Join(outDir, "will.pdf")
	if err := os.WriteFile(path, pdfBuffer, 0o644); err != nil {
		log.Fatalf("failed to write %s: %v", path, err)
	}
	log.Printf("wrote %s", path)
}

// buildRawPDF is the original builder for poa.pdf
func buildRawPDF(lines []string) []byte {
	escape := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, "(", `\(`)
		s = strings.ReplaceAll(s, ")", `\)`)
		return s
	}

	var content bytes.Buffer
	content.WriteString("BT\n/F1 11 Tf\n15 TL\n54 760 Td\n")
	for i, ln := range lines {
		content.WriteString("(" + escape(ln) + ") Tj\n")
		if i < len(lines)-1 {
			content.WriteString("T*\n")
		}
	}
	content.WriteString("ET")
	stream := content.String()

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] " +
			"/Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("%\xe2\xe3\xcf\xd3\n")

	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}

	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(objects)+1, xrefOffset)

	return buf.Bytes()
}
