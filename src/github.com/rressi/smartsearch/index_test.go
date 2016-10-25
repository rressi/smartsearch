package smartsearch

import (
	"bytes"
	"reflect"
	"testing"
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

	index, indexRaw, err := NewIndex(buf)
	if err != nil {
		t.Errorf("Cannot create index: %v", err)
	} else if index == nil {
		t.Error("Nil returned")
	} else if len(indexRaw) < 2 {
		t.Error("Invalid raw index returned")
	}

	var query string
	var expected_postings, postings []int

	query = "Text to test"
	expected_postings = []int{1, 2}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}

	query = "test/to-TEXT!"
	expected_postings = []int{1, 2}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}

	query = "test         to"
	expected_postings = []int{1, 2, 4}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query %v: postings=%v", query,
			postings)
	}

	query = "Th"
	expected_postings = []int{1, 2, 4}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query [%v]: postings=%v", query,
			postings)
	}

	query = "th "
	expected_postings = nil
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query [%v]: postings=%v", query,
			postings)
	}

	query = "-? "
	expected_postings = []int{1, 2, 3, 4}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query [%v]: postings=%v", query,
			postings)
	}

	query = ""
	expected_postings = []int{1, 2, 3, 4}
	postings, err = index.Search(query, -1)
	if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("Unexpected result with query [%v]: postings=%v", query,
			postings)
	}
}
