package smartsearch

import (
	"bytes"
	"sort"
	"strings"
)

func Tokenize(query string) (tokens []string) {
	var buf bytes.Buffer
	buf.ReadFrom(ReadNormalized(bytes.NewBufferString(query)))
	tokens = strings.Split(buf.String(), " ")
	sort.Strings(tokens)
	return tokens
}
