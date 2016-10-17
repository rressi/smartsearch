package smartsearch

import (
	"testing"
	"bytes"
	"reflect"
)

func TestIndex_Base(t *testing.T) {

	builder := NewIndexBuilder()
	builder.AddDocument(1, "This is a text to test something")
	builder.AddDocument(2, "This is another text to test something else")
	builder.AddDocument(3, "Now we would like to add another document")
	builder.AddDocument(4, "The more the better, we need to test!")

	buf := new(bytes.Buffer)
	builder.Dump(buf)
	if buf.Len() == 0 {
		t.Error("Cannot generate test binary")
	}

	index, err := NewIndex(buf)
	if err != nil {
		t.Errorf("Cannot create index: %v", err)
	} else if index == nil {
		t.Error("Nil returned")
	}

	var query string
	var expected_postings, postings []int

	query = "Text to test"
	expected_postings = []int{1, 2}
	postings, err = index.Search(query)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}

	query = "test/to-TEXT!"
	expected_postings = []int{1, 2}
	postings, err = index.Search(query)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}

	query = "test         to"
	expected_postings = []int{1, 2, 4}
	postings, err = index.Search(query)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}
}
