package smartsearch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

// A TrieBuilder is a tool that can be used to generate binary encoded tries
// for TrieReader.
type TrieBuilder interface {

	// It adds one term to the trie to build, and associates it to one posting.
	//
	// The passed posting should be strictly positive.
	//
	// If term is an empty string then the posting is added to the root node.
	Add(posting int, term string)

	// It adds many terms at once together with their postings.
	//
	// Data is passed as a map of terms -> list of postings.
	//
	// A value that is greater or equal to the maximum number of runes contained
	// inside the bulk data. For simplicity the maximum len of all the terms can
	// be used being it greater or equal than the number of runes.
	//
	// If a term contains an empty string then the posting is added to the root
	// node.
	AddBulk(bulkData map[string][]int, requiredRunes int)

	// Generates a trie and serializes to the passed io.Writer.
	//
	// It returns error on failures.
	Dump(dst io.Writer) error
}

// A TrieBuilder's node used internally by its implementation.
type trieNode struct {
	edges            map[rune]*trieNode
	postings         []int
	appendedPostings int
}

// It creates a TrieBuilder's node.
func newTrieNode() *trieNode {
	node := new(trieNode)
	node.edges = make(map[rune]*trieNode, 0)
	return node
}

// It implements TrieBuilder.Add
func (t *trieNode) Add(posting int, term string) {
	node := t
	if len(term) > 0 {
		for _, rune_ := range term {
			childNode, ok := node.edges[rune_]
			if !ok {
				childNode = newTrieNode()
				node.edges[rune_] = childNode
			}
			node = childNode
		}
	}
	node.postings = append(node.postings, posting)
	node.appendedPostings += 1
}

// Implementation of TrieBuilder.AddBulk
func (t *trieNode) AddBulk(bulkData map[string][]int, requiredRunes int) {

	// We can get some performance by iterating the terms in alphabetical order:
	var sortedTerms []string
	for term := range bulkData {
		sortedTerms = append(sortedTerms, term)
	}
	sort.Strings(sortedTerms)

	// We need a stack of nodes and a stack of runes in order to be able to
	// take advantage of the common prefixes that sorted terms have naturally:
	nodes := make([]*trieNode, requiredRunes+1)
	nodes[0] = t
	runes := make([]rune, requiredRunes+1)
	runes[0] = 0
	currPosition := 0

	// For each term in a sorted order:
	for _, term := range sortedTerms {

		// Walks to the target node starting from the last node sharing the
		// same prefix with last one previous:
		node := t
		for i, rune_ := range term {
			j := i + 1
			if j <= currPosition && runes[j] == rune_ {
				node = nodes[j] // Prefix match
			} else {
				currPosition = j // Prefix match stops here
				runes[j] = rune_
				var ok bool
				node, ok = nodes[i].edges[rune_]
				if !ok {
					node = newTrieNode()
					nodes[i].edges[rune_] = node
				}
				nodes[j] = node
			}
		}

		postings := bulkData[term]
		node.postings = append(node.postings, postings...)
		node.appendedPostings += len(postings)
	}
}

// It implements TrieBuilder.Dump
func (t *trieNode) Dump(dst io.Writer) error {
	_, err := t.dumpRec(dst)
	if err != nil {
		fmt.Errorf("trieNode.Dump: %v", err)
	}
	return err
}

// It recursively encodes one TrieBuilder's node.
func (t *trieNode) dumpRec(dst io.Writer) (sz int, err error) {

	// Utility function to save one value to a buffer:
	tmp := make([]byte, 16)
	writeInt := func(dst io.Writer, value int) (sz int, err error) {
		numBytes := binary.PutUvarint(tmp, uint64(value))
		return dst.Write(tmp[:numBytes])
	}

	// Consolidates collected postings:
	if len(t.postings) > 0 && t.appendedPostings > 0 {
		t.postings = SortDedupPostings(t.postings)
		t.appendedPostings = 0
	}

	var sz_ int

	// Dumps number of postings:
	sz_, err = writeInt(dst, len(t.postings))
	if err != nil {
		return
	}
	sz += sz_

	// Dumps number of edges:
	sz_, err = writeInt(dst, len(t.edges))
	if err != nil {
		return
	}
	sz += sz_

	// If any, dumps postings:
	if len(t.postings) > 0 {
		// Dumps the postings on a temporary buffer:
		encodedPostings := new(bytes.Buffer)
		_, err = t.dumpPostings(encodedPostings)

		// Dumps size of serialized posting buffer:
		sz_, err = writeInt(dst, encodedPostings.Len())
		if err != nil {
			return
		}
		sz += sz_

		// Dumps serialized posting buffer:
		sz_, err = dst.Write(encodedPostings.Bytes())
		if err != nil {
			return
		}
		sz += sz_
	}

	// If any, dumps all the edges and sub-nodes:
	if len(t.edges) > 0 {

		// Gets all the runes in sorted order:
		runes := make([]int, 0)
		for r := range t.edges {
			runes = append(runes, int(r))
		}
		sort.Ints(runes)

		// 2 temporary buffers for edges and child nodes:
		edgeBytes := new(bytes.Buffer)
		childNodeBytes := new(bytes.Buffer)

		// Dumps all the edges and their target nodes:
		previousRune := 0
		for _, rune_ := range runes {

			// Fetches and dumps the children node:
			childNode := t.edges[rune(rune_)]
			sz_, err = childNode.dumpRec(childNodeBytes)
			if err != nil {
				return
			}

			// Dumps edge's rune:
			_, err = writeInt(edgeBytes, rune_-previousRune)
			if err != nil {
				return
			}

			// Dumps size of serialized child node:
			_, err = writeInt(edgeBytes, sz_)
			if err != nil {
				return
			}

			previousRune = rune_
		}

		// Dumps size of serialized edges:
		sz_, err = writeInt(dst, edgeBytes.Len())
		if err != nil {
			return
		}
		sz += sz_

		// Dumps edges:
		sz_, err = dst.Write(edgeBytes.Bytes())
		if err != nil {
			return
		}
		sz += sz_

		// Dumps sub nodes:
		sz_, err = dst.Write(childNodeBytes.Bytes())
		if err != nil {
			return
		}
		sz += sz_
	}

	return
}

// It encodes all the postings associated to one TrieBuilder's node.
func (t *trieNode) dumpPostings(dst io.Writer) (sz int, err error) {

	if len(t.postings) == 0 {
		return
	}

	// Consolidates collected postings:
	if t.appendedPostings > 0 {
		t.postings = SortDedupPostings(t.postings)
		t.appendedPostings = 0
	}

	// Dumps all the postings:
	var sz_ int
	previousPosting := 0
	tmp := make([]byte, 16)
	for _, posting := range t.postings {

		// Serializes the increment of current posting:
		numBytes := binary.PutUvarint(tmp, uint64(posting-previousPosting))

		// Dumps it to the target writer:
		sz_, err = dst.Write(tmp[:numBytes])
		if err != nil {
			return
		}
		sz += sz_

		previousPosting = posting
	}

	return
}

// -----------------------------------------------------------------------------

// Creates a new TrieBuilder.
func NewTrieBuilder() TrieBuilder {
	return newTrieNode()
}
