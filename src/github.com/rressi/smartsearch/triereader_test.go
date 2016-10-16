package smartsearch

import (
	"io"
	"reflect"
	"testing"
)

func TestTrieReader_Empty(t *testing.T) {

	// builder := NewTrieBuilder()

	source_bytes := []byte{0, 0}
	expected_node := Node{0, 0, 2, 2}
	expected_posting := 0
	expected_edge := Edge{0, 0}

	reader, node, err := NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("NewTrieReader failed: %v", err)
	} else if reader == nil {
		t.Errorf("NewTrieReader failed: reader=%v", reader)
	} else if node != expected_node {
		t.Errorf("NewTrieReader failed: node=%v", node)
	}

	posting, err := reader.ReadPosting()
	if err != io.EOF {
		t.Errorf("ReadPosting failed: err=%v", err)
	} else if posting != expected_posting {
		t.Errorf("ReadPosting failed: posting=%v", posting)
	}

	edge, err := reader.ReadEdge()
	if err != io.EOF {
		t.Errorf("ReadEdge failed: err=%v", err)
	} else if edge != expected_edge {
		t.Errorf("ReadEdge failed: edge=%v", edge)
	}
}

func TestTrieReader_VoidTerm(t *testing.T) {

	// builder := NewTrieBuilder()
	// builder.Add(1, "")
	// builder.Add(2, "")

	source_bytes := []byte{2, 0, 2, 1, 1}
	expected_node := Node{2, 0, 3, 5}
	expected_postings := []int{1, 2}
	expected_edge := Edge{0, 0}

	// Creates a reader, it points to the root node:
	reader, node, err := NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("NewTrieReader failed: %v", err)
	} else if reader == nil {
		t.Errorf("NewTrieReader failed: reader=%v", reader)
	} else if node != expected_node {
		t.Errorf("NewTrieReader failed: node=%v", node)
	}

	// Fetches all the postings:
	for i, expected_posting := range expected_postings {
		posting, err := reader.ReadPosting()
		if err != nil {
			t.Errorf("ReadPosting failed: i=%v, err=%v", i, err)
		} else if posting != expected_posting {
			t.Errorf("ReadPosting failed: posting[%i]=%v", i, posting)
		}
	}

	// Should return EOF asking one more:
	posting, err := reader.ReadPosting()
	if err != io.EOF {
		t.Errorf("ReadPosting failed: err=%v", err)
	} else if posting != 0 {
		t.Errorf("ReadPosting failed: posting=%v", posting)
	}

	// Restarts the game:
	node, err = reader.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	} else if node != expected_node {
		t.Errorf("Reset failed: node=%v", node)
	}

	// Should return all the postings at once:
	postings, err := reader.ReadAllPostings()
	if err != nil {
		t.Errorf("ReadAllPostings failed: err=%v", err)
	} else if !reflect.DeepEqual(postings, expected_postings) {
		t.Errorf("ReadAllPostings failed: postings=%v", postings)
	}

	// Should return EOF asking one edge:
	edge, err := reader.ReadEdge()
	if err != io.EOF {
		t.Errorf("ReadEdge failed: err=%v", err)
	} else if edge != expected_edge {
		t.Errorf("ReadEdge failed: edge=%v", edge)
	}
}

func TestTrieReader_Base(t *testing.T) {

	// builder := NewTrieBuilder()
	// builder.Add(1, "A")
	// builder.Add(2, "A")
	// builder.Add(1, "B")
	// builder.Add(2, "B")

	source_bytes := []byte{0, 2, 4, 65, 5, 1, 5, 2, 0, 2, 1, 1, 2, 0, 2, 1, 1}
	expected_nodes := []Node{{0, 2, 2, 2}, {2, 0, 10, 12}, {2, 0, 15, 17}}
	expected_edges := []Edge{{rune('A'), 7}, {rune('B'), 12}}
	expected_edge_eof := Edge{0, 0}
	expected_postings := []int{1, 2, 1, 2}
	expected_posting_eof := 0

	// Creates a reader, it points to the root node:
	reader, node, err := NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("NewTrieReader failed: %v", err)
	} else if reader == nil {
		t.Errorf("NewTrieReader failed: reader=%v", reader)
	} else if node != expected_nodes[0] {
		t.Errorf("NewTrieReader failed: node[0]=%v", node)
	}

	// Should return EOF asking one posting:
	posting, err := reader.ReadPosting()
	if err != io.EOF {
		t.Errorf("ReadPosting failed: err=%v", err)
	} else if posting != expected_posting_eof {
		t.Errorf("ReadPosting failed: posting=%v", posting)
	}

	for i, expected_edge := range expected_edges {
		edge, err := reader.ReadEdge()
		if err != nil {
			t.Errorf("ReadEdge failed: err=%v", err)
		} else if edge != expected_edge {
			t.Errorf("ReadEdge failed: edge[%v]=%v", i, edge)
		}
	}

	// Should return EOF asking one edge more:
	edge, err := reader.ReadEdge()
	if err != io.EOF {
		t.Errorf("ReadEdge failed: err=%v", err)
	} else if edge != expected_edge_eof {
		t.Errorf("ReadEdge failed: edge=%v", edge)
	}

	// Restarts the game:
	node, err = reader.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	} else if node != expected_nodes[0] {
		t.Errorf("Reset failed: node[0]=%v", node)
	}

	// Should return all the edges at once:
	edges, err := reader.ReadAllEdges()
	if err != nil {
		t.Errorf("ReadAllEdges failed: err=%v", err)
	} else if !reflect.DeepEqual(edges, expected_edges) {
		t.Errorf("ReadAllEdges failed: edges=%v", edges)
	}

	j := 0
	for i, edge := range edges {
		node, err = reader.EnterNode(edge)
		if err != nil {
			t.Errorf("EnterNode failed: %v", err)
		} else if node != expected_nodes[i+1] {
			t.Errorf("EnterNode failed: node[%v]=%v", i, node)
		}

		// Should return all the postings at once:
		postings, err := reader.ReadAllPostings()
		if err != nil {
			t.Errorf("ReadAllPostings failed: err=%v", err)
		} else if !reflect.DeepEqual(postings,
			expected_postings[j:j+node.NumPostings]) {
			t.Errorf("ReadAllPostings failed: postings=%v", postings)
		}
		j += node.NumPostings

		// Should return EOF asking one posting more:
		posting, err := reader.ReadPosting()
		if err != io.EOF {
			t.Errorf("ReadPosting failed: err=%v", err)
		} else if posting != expected_posting_eof {
			t.Errorf("ReadPosting failed: posting=%v", posting)
		}

		// Should return EOF asking one edge:
		edge, err := reader.ReadEdge()
		if err != io.EOF {
			t.Errorf("ReadEdge failed: err=%v", err)
		} else if edge != expected_edge_eof {
			t.Errorf("ReadEdge failed: edge=%v", edge)
		}
	}

}

func TestTrieReader_Match(t *testing.T) {

	var err error
	var source_bytes []byte
	var reader *TrieReader
	var terms []string
	var expected_matches []Node

	// builder := NewTrieBuilder()
	source_bytes = []byte{0, 0}
	terms = []string{"", "A", "B", "ABC"}
	expected_matches = []Node{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0}}
	reader, _, err = NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("Cannot create trie reader from bytes: %v", source_bytes)
	}
	for i, term := range terms {
		node, err := reader.Match(term)
		if err != nil {
			t.Errorf("Cannot match term '%v': %v", term, source_bytes)
		} else if node != expected_matches[i] {
			t.Errorf("Unexpected match: term='%v' node[%d]=%v", term, i, node)
		}
	}

	// builder := NewTrieBuilder()
	// builder.Add(1, "A")
	// builder.Add(2, "A")
	// builder.Add(1, "B")
	// builder.Add(2, "B")
	source_bytes = []byte{0, 2, 4, 65, 5, 1, 5, 2, 0, 2, 1, 1, 2, 0, 2, 1, 1}
	terms = []string{"", "A", "B", "ABC"}
	expected_matches = []Node{
		{0, 0, 0, 0},
		{2, 0, 10, 12},
		{2, 0, 15, 17},
		{0, 0, 0, 0}}
	reader, _, err = NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("Cannot create trie reader from bytes: %v", source_bytes)
	}
	for i, term := range terms {
		reader.Reset()
		node, err := reader.Match(term)
		if err != nil {
			t.Errorf("Cannot match term '%v': %v", term, source_bytes)
		} else if node != expected_matches[i] {
			t.Errorf("Unexpected match: term='%v' node[%d]=%v", term, i, node)
		}
	}

	// builder := NewTrieBuilder()
	// builder.Add(1, "ABC")
	// builder.Add(2, "BCA")
	// builder.Add(3, "CAB")
	source_bytes = []byte{0, 3, 6, 65, 14, 1, 14, 1, 14, 0, 1, 2, 66, 9, 0, 1,
		2, 67, 4, 1, 0, 1, 1, 0, 1, 2, 67, 9, 0, 1, 2, 65, 4, 1, 0, 1, 2, 0,
		1, 2, 65, 9, 0, 1, 2, 66, 4, 1, 0, 1, 3}
	terms = []string{"", "A", "BC", "CAB", "AA", "CBA"}
	expected_matches = []Node{
		{0, 0, 0, 0},
		{0, 1, 11, 11},
		{0, 1, 30, 30},
		{1, 0, 50, 51},
		{0, 0, 0, 0},
		{0, 0, 0, 0}}
	reader, _, err = NewTrieReader(source_bytes)
	if err != nil {
		t.Errorf("Cannot create trie reader from bytes: %v", source_bytes)
	}
	for i, term := range terms {
		reader.Reset()
		node, err := reader.Match(term)
		if err != nil {
			t.Errorf("Cannot match term '%v': %v", term, source_bytes)
		} else if node != expected_matches[i] {
			t.Errorf("Unexpected match: term='%v' node[%d]=%v", term, i, node)
		}
	}
}
