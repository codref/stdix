package search

import (
	"github.com/codref/stdix/internal/registry"
)

// IndexTerms builds a weighted, pre-tokenised term list for a standard.
// Field weights are achieved by repetition:
//
//	language    × 5
//	tags        × 3
//	applies_when × 3
//	title       × 2
//	rules       × 1
func IndexTerms(s registry.Standard) []string {
	var terms []string

	// Language — 5×
	if s.Language != "" {
		lang := Tokenize(s.Language)
		for i := 0; i < 5; i++ {
			terms = append(terms, lang...)
		}
	}

	// Tags — 3×
	for _, tag := range s.Tags {
		tok := Tokenize(tag)
		for i := 0; i < 3; i++ {
			terms = append(terms, tok...)
		}
	}

	// AppliesWhen — 3×
	for _, phrase := range s.AppliesWhen {
		tok := Tokenize(phrase)
		for i := 0; i < 3; i++ {
			terms = append(terms, tok...)
		}
	}

	// Title — 2×
	for _, t := range Tokenize(s.Title) {
		terms = append(terms, t, t)
	}

	// Rules — 1×
	for _, rule := range s.Rules {
		terms = append(terms, Tokenize(rule)...)
	}

	return terms
}
