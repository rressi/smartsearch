package smartsearch

import (
	"bytes"
	"sort"
	"strings"
)

// Given a free text, produces normalized tokens.
//
// If last character of the passed input was a valid one then considers as
// potentially incomplete it and returns it apart.
//
// It returns:
// - all complete tokens, sorted and deduplicated.
// - optionally the las token a part if it was considered to be potentially
//   incomplete.
func TokenizeForSearch(query string) (tokens []string,
	incomplete_token string) {

	if len(query) == 0 {
		return // Sorry, no tokens found.
	}

	var _tokens []string
	var _incomplete_token string

	// Normalizes the query:
	var buf bytes.Buffer
	buf.ReadFrom(ReadNormalized(bytes.NewBufferString(query)))
	normalized_query := buf.String()
	if normalized_query == "" || normalized_query == " " {
		return // Sorry, no tokens found.
	}

	// Extracts all non-empty tokens:
	for _, token := range strings.Split(normalized_query, " ") {
		if len(token) > 0 {
			_tokens = append(_tokens, token)
		}
	}
	if len(_tokens) == 0 {
		return // Sorry, no tokens found.
	}

	// If we don't have a separator at the end of the query means that the last
	// typed character may be part of a term the user is still writing:
	if normalized_query[len(normalized_query)-1] != ' ' {
		_incomplete_token = _tokens[len(_tokens)-1]
		_tokens = _tokens[:len(_tokens)-1]
	}

	// Sorts and deduplicates extracted tokens:
	if len(_tokens) > 1 {
		sort.Strings(_tokens)
		i := 0
		for j := 1; j < len(_tokens); j++ {
			if _tokens[i] != _tokens[j] {
				i++
				_tokens[i] = _tokens[j]
			}
		}
		_tokens = _tokens[:i+1]
	}

	// Generates the final result:
	if len(_tokens) > 0 {
		tokens = make([]string, len(_tokens))
		copy(tokens, _tokens)
	}
	incomplete_token = _incomplete_token

	return
}
