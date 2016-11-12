package smartsearch

import (
	"reflect"
	"testing"
)

func TestIndexer_Base(t *testing.T) {

	contentA := "YES!-This ìs ä fÄncy, is a string"
	contentB := "This ìs à book"
	expected_terms := IndexedTerms{
		{"a", []int{10, 12}, 3},
		{"book", []int{12}, 1},
		{"fancy", []int{10}, 1},
		{"is", []int{10, 12}, 3},
		{"string", []int{10}, 1},
		{"this", []int{10, 12}, 2},
		{"yes", []int{10}, 1}}

	indexer := NewIndexer()
	indexer.AddContent(10, []byte(contentA))
	indexer.AddContent(12, []byte(contentB))
	indexer.Finish()
	terms, err := indexer.Result()
	if err != nil {
		t.Errorf("Indexer failed: %v", err)
	} else if !reflect.DeepEqual(terms, expected_terms) {
		t.Errorf("Unexpected result: terms=%v", terms)
	}
}
