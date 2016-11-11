package smartsearch

import (
	"strings"
)

type IndexedTerm struct {
	term        string
	postings    []int
	occurrences int
}

type IndexedTerms []IndexedTerm

// Implementation of sort.Interface
func (s IndexedTerms) Len() int {
	return len(s)
}

// Implementation of sort.Interface
func (s IndexedTerms) Less(i, j int) bool {
	return strings.Compare(s[i].term, s[j].term) < 0
}

// Implementation of sort.Interface
func (s IndexedTerms) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
