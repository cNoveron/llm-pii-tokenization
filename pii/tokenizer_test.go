package pii

import "testing"

func TestTokenizeRoundTrip(t *testing.T) {
	input := "John Smith (SSN 123-45-6789, born 04/12/1950) leaves $500,000.00 " +
		"to Emily Smith. Contact john.smith@example.com or (415) 555-0198."

	tokenized, tokenMap := Tokenize(input)

	// No raw PII must remain in the tokenized text.
	for _, leaked := range []string{
		"John Smith", "Emily Smith", "123-45-6789", "04/12/1950",
		"$500,000.00", "john.smith@example.com", "555-0198",
	} {
		if contains(tokenized, leaked) {
			t.Errorf("tokenized text still contains raw PII %q: %s", leaked, tokenized)
		}
	}

	// Every expected token type should have been produced.
	for _, tok := range []string{"{{PERSON_1}}", "{{SSN_1}}", "{{DOB_1}}", "{{AMOUNT_1}}", "{{EMAIL_1}}", "{{PHONE_1}}"} {
		if _, ok := tokenMap[tok]; !ok {
			t.Errorf("expected token %q not found in map %v", tok, tokenMap)
		}
	}

	// Round-trip restores the original.
	if got := Detokenize(tokenized, tokenMap); got != input {
		t.Errorf("round-trip mismatch:\n got:  %s\n want: %s", got, input)
	}
}

func TestNewFromNoCollision(t *testing.T) {
	// A document that already used PERSON_1 for John Smith.
	docText, docMap := Tokenize("John Smith is the testator.")

	// A follow-up question mentioning a *different* person plus the same one.
	tk := NewFrom(docMap)
	qText := tk.Apply("Does Sarah Chen work with John Smith?")
	combined := tk.Map()

	// The known document entity must reuse its existing token.
	if !contains(qText, "{{PERSON_1}}") {
		t.Errorf("John Smith should reuse {{PERSON_1}}; got %q", qText)
	}
	// The new entity must get a fresh, non-colliding token.
	if !contains(qText, "{{PERSON_2}}") {
		t.Errorf("Sarah Chen should get {{PERSON_2}}; got %q", qText)
	}
	if combined["{{PERSON_1}}"] != "John Smith" {
		t.Errorf("PERSON_1 should still map to John Smith; got %q", combined["{{PERSON_1}}"])
	}
	if combined["{{PERSON_2}}"] != "Sarah Chen" {
		t.Errorf("PERSON_2 should map to Sarah Chen; got %q", combined["{{PERSON_2}}"])
	}
	_ = docText
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
