package smartsearch

import (
	"bytes"
	"sort"
	"strings"
)

func Tokenize(query string) (tokens []string) {

	if len(query) == 0 {
		return
	}

	// Normalizes the query:
	var buf bytes.Buffer
	buf.ReadFrom(ReadNormalized(bytes.NewBufferString(query)))

	// Extracts all non-empty tokens:
	for _, token := range strings.Split(buf.String(), " ") {
		if len(token) > 0 {
			tokens = append(tokens, token)
		}
	}

	// Sorts and deduplicates extracted tokens:
	if len(tokens) > 1 {
		sort.Strings(tokens)
		i := 0
		for j := 1; j < len(tokens); j++ {
			if tokens[i] != tokens[j] {
				i++
				tokens[i] = tokens[j]
			}
		}
		tokens = tokens[:i+1]
	}

	return
}
