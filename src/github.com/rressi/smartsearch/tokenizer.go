package smartsearch

import (
	"sort"
	"strings"
)

type Tokenizer interface {
	Apply(query string) (tokens []string)
	ForSearch(query string) (tokens []string, incompleteTerm string)
}

type tokenizerImpl struct {
	normalizer Normalizer
}

func NewTokenizer() Tokenizer {
	tokenizer := new(tokenizerImpl)
	tokenizer.normalizer = NewNormalizer()
	return tokenizer
}

// Given a free text, produces normalized tokens.
//
// It returns:
// - extracted tokens in the same original order.
func (t *tokenizerImpl) Apply(query string) (tokens []string) {

	if len(query) == 0 {
		return // Sorry, no tokens found.
	}

	// Normalizes the query:
	normalized_query := t.normalizer.Apply(query)
	if normalized_query == "" || normalized_query == " " {
		return // Sorry, no tokens found.
	}

	// Extracts all non-empty tokens:
	for _, token := range strings.Split(normalized_query, " ") {
		if len(token) > 0 {
			tokens = append(tokens, token)
		}
	}

	return
}

// Given a free text, produces normalized tokens.
//
// If last character of the passed input was a valid one then considers as
// potentially incomplete it and returns it apart.
//
// It returns:
// - all complete tokens, sorted and deduplicated.
// - optionally the las token a part if it was considered to be potentially
//   incomplete.
func (t *tokenizerImpl) ForSearch(query string) (tokens []string,
	incompleteToken string) {

	if len(query) == 0 {
		return // Sorry, no tokens found.
	}

	// Normalizes the query:
	normalized_query := t.normalizer.Apply(query)
	if normalized_query == "" || normalized_query == " " {
		return // Sorry, no tokens found.
	}

	// Extracts all non-empty tokens:
	var tokens_ []string
	for _, token := range strings.Split(normalized_query, " ") {
		if len(token) > 0 {
			tokens_ = append(tokens_, token)
		}
	}
	if len(tokens_) == 0 {
		return // Sorry, no tokens found.
	}

	// If we don't have a separator at the end of the query means that the last
	// typed character may be part of a term the user is still writing:
	var incompleteToken_ string
	if normalized_query[len(normalized_query)-1] != ' ' {
		incompleteToken_ = tokens_[len(tokens_)-1]
		tokens_ = tokens_[:len(tokens_)-1]
	}

	// Sorts and deduplicates extracted tokens:
	if len(tokens_) > 1 {
		sort.Strings(tokens_)
		i := 0
		for j := 1; j < len(tokens_); j++ {
			if tokens_[i] != tokens_[j] {
				i++
				tokens_[i] = tokens_[j]
			}
		}
		tokens_ = tokens_[:i+1]
	}

	// Generates the final result:
	tokens = append(tokens, tokens_...)
	incompleteToken = incompleteToken_
	return
}
