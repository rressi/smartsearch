package smartsearch

import (
	"bytes"
	"reflect"
	"testing"
	"io"
)

func TestIndexBuilder_AddDocument(t *testing.T) {

	var err error

	builder := NewIndexBuilder()
	if builder == nil {
		t.Error("Cannot create index builder")
	}
	builder.AddDocument(1, "The lazy fox is running fast")
	builder.AddDocument(2, "But what does man you don't know!")
	builder.AddDocument(3, "A frog jumps on the table, funny frog ")

	indexBytes := new(bytes.Buffer)
	err = builder.Dump(indexBytes)
	if err != nil {
		t.Errorf("Cannot dump index: %v", err)
	}

	reader, node, err := NewTrieReader(indexBytes.Bytes())
	if err != nil {
		t.Errorf("Cannot create index reader: %v", err)
	} else if reader == nil {
		t.Error("Cannot create index reader")
	} else if node.NumEdges == 0 && node.NumPostings == 0 {
		t.Errorf("Unexpected empty index opened: %v", node)
	}

	terms := []string{"frog", "the", "but", "don", "lazy", "unknows", ""}
	all_expected_postings := [][]int{{3}, {1, 3}, {2}, {2}, {1}, {}, {}}
	for i, term := range terms {
		reader.Reset()

		node, err = reader.Match(term)
		if err != nil {
			t.Errorf("Failure while matching: %v", err)
		}

		expected_postings := all_expected_postings[i]
		if len(expected_postings) != node.NumPostings {
			t.Errorf("Unexpected number of postings for term '%v': %v", term,
				node.NumPostings)
		}

		var postings []int
		postings, err = reader.ReadAllPostings()
		if node.NumPostings == 0 {
			if err != io.EOF {
				t.Errorf("Io was expected while matching '%v'", term)
			}
		} else {
			if err != nil {
				t.Errorf("Failure while matching term '%v': %v", term, err)
			} else if !reflect.DeepEqual(postings, expected_postings) {
				t.Errorf("Unexpected postings for term '%v': %v", term,
					postings)
			}
		}
	}
}
