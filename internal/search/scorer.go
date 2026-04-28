package search

import (
	"sort"
	"strings"

	"github.com/stdix/stdix/internal/db"
)

// LangBonus is added to the BM25 score when the query language matches a standard.
const LangBonus = 50.0

// Result is a scored standard with an optional match explanation.
type Result struct {
	Standard   db.IndexedStandard
	Score      float64
	MatchedWhy string // first applies_when phrase that shares a token with the query
}

// Score ranks standards against query text, applying an optional language bonus.
// lang may be empty to skip the bonus.
func Score(query, lang string, standards []db.IndexedStandard) []Result {
	queryTerms := Tokenize(query)

	docs := make([][]string, len(standards))
	for i, s := range standards {
		docs[i] = s.Terms
	}
	c := newCorpus(docs)

	results := make([]Result, 0, len(standards))
	for i, s := range standards {
		sc := c.score(i, queryTerms)
		if lang != "" && strings.EqualFold(s.Language, lang) {
			sc += LangBonus
		}
		matchedWhy := firstMatchedPhrase(queryTerms, s.AppliesWhen)
		results = append(results, Result{
			Standard:   s,
			Score:      sc,
			MatchedWhy: matchedWhy,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

// firstMatchedPhrase returns the first applies_when phrase that shares at least
// one token with queryTerms, or empty string if none match.
func firstMatchedPhrase(queryTerms []string, phrases []string) string {
	qt := map[string]bool{}
	for _, t := range queryTerms {
		qt[t] = true
	}
	for _, phrase := range phrases {
		for _, t := range Tokenize(phrase) {
			if qt[t] {
				return phrase
			}
		}
	}
	return ""
}
