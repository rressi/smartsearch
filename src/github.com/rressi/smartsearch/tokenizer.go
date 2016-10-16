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

	var buf bytes.Buffer
	buf.ReadFrom(ReadNormalized(bytes.NewBufferString(query)))

	for _, token := range strings.Split(buf.String(), " ") {
		if len(token) > 0 {
			tokens = append(tokens, token)
		}
	}
	sort.Strings(tokens)

	return
}
