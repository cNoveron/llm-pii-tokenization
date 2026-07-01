// Package pii provides mock, regex-based PII detection for the PoC. It is
// deliberately simple: entities are matched with regular expressions (and, for
// names, a hardcoded list), replaced with stable per-type tokens, and can be
// restored via the token map. A production system would use a real detector
// (Presidio, GCP DLP, an NER model), not this.
package pii

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Entity regexes. Detection order matters (see entities below): more specific /
// consuming patterns run first so a value isn't half-claimed by a later one
// (e.g. EMAIL runs before PERSON so a name inside an address becomes one email
// token).
var (
	ssnRe    = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	dobRe    = regexp.MustCompile(`\b(?:0?[1-9]|1[0-2])/\d{1,2}/\d{4}\b`)
	amountRe = regexp.MustCompile(`\$[\d,]+(?:\.\d{2})?`)
	phoneRe  = regexp.MustCompile(`\(?\d{3}\)?[-.\s]\d{3}[-.\s]\d{4}`)
	emailRe  = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	personRe *regexp.Regexp // built in init from knownNames
)

// knownNames are the full names present in the sample documents. This list is
// the mock stand-in for real name detection and MUST stay in sync with the
// names embedded by the sample generator (sample/gen/main.go).
var knownNames = []string{
	"John Smith",
	"Margaret Smith",
	"Emily Smith",
	"Thomas Smith",
	"Robert Chen",
	"Margaret Whitfield",
	"David Whitfield",
	"Sarah Chen",
}

// tokenRe parses a token back into its (type, number) parts, used when seeding
// a tokenizer from an existing map.
var tokenRe = regexp.MustCompile(`^\{\{([A-Z]+)_(\d+)\}\}$`)

type entity struct {
	name string // token prefix, e.g. "PERSON"
	re   *regexp.Regexp
}

var entities []entity

func init() {
	// Build the PERSON alternation longest-first so a longer full name wins
	// over any shorter name it might share a prefix with.
	names := make([]string, len(knownNames))
	copy(names, knownNames)
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = regexp.QuoteMeta(n)
	}
	personRe = regexp.MustCompile(`\b(?:` + strings.Join(quoted, "|") + `)\b`)

	entities = []entity{
		{"SSN", ssnRe},
		{"DOB", dobRe},
		{"AMOUNT", amountRe},
		{"PHONE", phoneRe},
		{"EMAIL", emailRe},
		{"PERSON", personRe},
	}
}

// Tokenizer replaces PII with stable tokens. Values seen more than once reuse
// the same token; each entity type has its own counter. A Tokenizer can be
// seeded from an existing token map (NewFrom) so a follow-up text — e.g. a Q&A
// question — extends the same token space instead of colliding with it.
type Tokenizer struct {
	tokenMap     map[string]string // token -> original value
	valueToToken map[string]string // "TYPE|value" -> token
	counters     map[string]int    // TYPE -> highest number used
}

// New returns an empty tokenizer.
func New() *Tokenizer {
	return &Tokenizer{
		tokenMap:     make(map[string]string),
		valueToToken: make(map[string]string),
		counters:     make(map[string]int),
	}
}

// NewFrom returns a tokenizer seeded from an existing token->value map so that
// Apply reuses existing tokens for known values and continues numbering for new
// ones. The passed map is not modified.
func NewFrom(existing map[string]string) *Tokenizer {
	t := New()
	for token, value := range existing {
		t.tokenMap[token] = value
		m := tokenRe.FindStringSubmatch(token)
		if m == nil {
			continue
		}
		typ, num := m[1], m[2]
		t.valueToToken[typ+"|"+value] = token
		if n, err := strconv.Atoi(num); err == nil && n > t.counters[typ] {
			t.counters[typ] = n
		}
	}
	return t
}

// Apply replaces all detected PII in text with tokens and returns the
// tokenized text. It mutates the tokenizer's internal map.
func (t *Tokenizer) Apply(text string) string {
	for _, e := range entities {
		text = e.re.ReplaceAllStringFunc(text, func(match string) string {
			key := e.name + "|" + match
			if token, ok := t.valueToToken[key]; ok {
				return token
			}
			t.counters[e.name]++
			token := fmt.Sprintf("{{%s_%d}}", e.name, t.counters[e.name])
			t.valueToToken[key] = token
			t.tokenMap[token] = match
			return token
		})
	}
	return text
}

// Map returns a copy of the accumulated token->value map.
func (t *Tokenizer) Map() map[string]string {
	out := make(map[string]string, len(t.tokenMap))
	for k, v := range t.tokenMap {
		out[k] = v
	}
	return out
}

// Tokenize detects and replaces PII in text, returning the tokenized text and
// the token->original map. It is the convenience form of New().Apply().
func Tokenize(text string) (tokenized string, tokenMap map[string]string) {
	t := New()
	tokenized = t.Apply(text)
	return tokenized, t.Map()
}

// Detokenize restores original values by replacing every token in tokenMap.
// Tokens are replaced longest-first as a defensive measure (the "}}" terminator
// already prevents prefix ambiguity between e.g. {{PERSON_1}} and {{PERSON_10}}).
func Detokenize(text string, tokenMap map[string]string) string {
	tokens := make([]string, 0, len(tokenMap))
	for token := range tokenMap {
		tokens = append(tokens, token)
	}
	sort.Slice(tokens, func(i, j int) bool { return len(tokens[i]) > len(tokens[j]) })
	for _, token := range tokens {
		text = strings.ReplaceAll(text, token, tokenMap[token])
	}
	return text
}
