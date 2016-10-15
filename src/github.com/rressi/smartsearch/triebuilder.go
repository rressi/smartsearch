package smartsearch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type TrieBuilder interface {
	Add(posting int, term string)
	Dump(dst io.Writer) error
}

type trieNode struct {
	edges    map[rune]*trieNode
	postings map[int]int
}

func newTrieNode() *trieNode {
	node := new(trieNode)
	node.edges = make(map[rune]*trieNode, 0)
	node.postings = make(map[int]int, 0)
	return node
}

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
	node.postings[posting] += 1
}

func (t *trieNode) Dump(dst io.Writer) error {
	_, err := t.dumpRec(dst)
	if err != nil {
		fmt.Errorf("trieNode.Dump: %v", err)
	}
	return err
}

func (t *trieNode) dumpRec(dst io.Writer) (sz int, err error) {

	// Utility function to save one value to a buffer:
	tmp := make([]byte, 16)
	writeInt := func(dst io.Writer, value int) (sz int, err error) {
		numBytes := binary.PutUvarint(tmp, uint64(value))
		return dst.Write(tmp[:numBytes])
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

func (t *trieNode) dumpPostings(dst io.Writer) (sz int, err error) {

	// It fetches all the postings and sorts them:
	postings := make([]int, len(t.postings))
	i := 0
	for posting := range t.postings {
		postings[i] = posting
		i++
	}
	sort.Ints(postings) // NOTE: they are already deduplicated.

	// Dumps all the postings:
	var sz_ int
	previousPosting := 0
	tmp := make([]byte, 16)
	for _, posting := range postings {

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

func NewTrieBuilder() TrieBuilder {
	root := newTrieNode()
	return root
}
