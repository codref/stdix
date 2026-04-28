package search

import (
	"math"
	"strings"
	"unicode"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

// Tokenize splits text into lowercase alphanumeric tokens.
func Tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// corpus holds precomputed BM25 statistics for a set of documents.
type corpus struct {
	n     int            // total number of documents
	avgDL float64        // average document length (in tokens)
	df    map[string]int // document frequency: term → count of docs containing it
	docs  [][]string     // tokenised documents
}

// newCorpus builds a corpus from a slice of pre-tokenised documents.
func newCorpus(docs [][]string) *corpus {
	df := map[string]int{}
	total := 0
	for _, doc := range docs {
		total += len(doc)
		seen := map[string]bool{}
		for _, t := range doc {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	avgDL := 0.0
	if len(docs) > 0 {
		avgDL = float64(total) / float64(len(docs))
	}
	return &corpus{n: len(docs), avgDL: avgDL, df: df, docs: docs}
}

// score computes the BM25 score for document docIdx against queryTerms.
func (c *corpus) score(docIdx int, queryTerms []string) float64 {
	doc := c.docs[docIdx]
	tf := map[string]int{}
	for _, t := range doc {
		tf[t]++
	}
	dl := float64(len(doc))
	var total float64
	for _, term := range queryTerms {
		f := float64(tf[term])
		if f == 0 {
			continue
		}
		idf := math.Log(
			(float64(c.n)-float64(c.df[term])+0.5)/
				(float64(c.df[term])+0.5) + 1,
		)
		numerator := f * (bm25K1 + 1)
		denominator := f + bm25K1*(1-bm25B+bm25B*dl/c.avgDL)
		total += idf * numerator / denominator
	}
	return total
}
