package smartsearch

import (
	"bytes"
	"encoding/binary"
	"sort"
)

type TrieBuilder interface {
	Add(posting int, term string)
	Dump() []byte
}

type trieNode struct {
	edges    map[rune]*trieNode
	postings map[int]int
}

func (t *trieNode) Add(posting int, term string) {
	node := t
	for i, codePoint := range term {
		node, ok := node.edges[codePoint]
		if !ok {
			newNode := newTrieNode()
			node.edges[codePoint] = newNode
			node = newNode
		}
		if i+1 == len(term) {
			node.postings[posting] += 1
		}
	}
}

func (t *trieNode) Dump() []byte {
	buf := new(bytes.Buffer)
	t.dumpRecursion(buf)
	return buf.Bytes()
}

func (t *trieNode) dumpRecursion(targetBuffer *bytes.Buffer) int {

	startOffset := targetBuffer.Len()

	// Utility function to save one UVarint to a buffer:
	tmp := make([]byte, 16)
	writeInt := func(dst *bytes.Buffer, value int) {
		numBytes := binary.PutUvarint(tmp, uint64(value))
		dst.Write(tmp[:numBytes])
	}

	// Dumps number of postings and number of edges:
	writeInt(targetBuffer, len(t.postings))
	writeInt(targetBuffer, len(t.edges))

	// Dumps the postings:
	if len(t.postings) > 0 {
		encodedPostings := t.dumpPostings()
		writeInt(targetBuffer, len(encodedPostings))
		targetBuffer.Write(encodedPostings)
	} else {
		writeInt(targetBuffer, 0)
	}

	// Dumps all the edges and sub-nodes:
	if len(t.edges) > 0 {

		// Gets all the runes in sorted order:
		runes := make([]int, 0)
		for r := range runes {
			runes = append(runes, int(r))
		}
		sort.Ints(runes)

		// 2 temporary buffers for edges and child nodes:
		childNodeBytes := new(bytes.Buffer)
		edgeBytes := new(bytes.Buffer)

		// Dumps all the edges and their target nodes:
		for _, rune_ := range runes {
			childNode := t.edges[rune(rune_)]
			numChildBytes := childNode.dumpRecursion(childNodeBytes)
			writeInt(edgeBytes, rune_)
			writeInt(edgeBytes, numChildBytes)
		}

		writeInt(targetBuffer, edgeBytes.Len())
		targetBuffer.Write(edgeBytes.Bytes())
		targetBuffer.Write(childNodeBytes.Bytes())
	}

	return targetBuffer.Len() - startOffset
}

func (t *trieNode) dumpPostings() []byte {

	buf := new(bytes.Buffer)

	postings := make([]int, len(t.postings))
	i := 0
	for posting := range t.postings {
		postings[i] = posting
		i++
	}
	sort.Ints(postings)

	previousPosting := 0
	tmp := make([]byte, 16)
	for _, posting := range postings {
		numBytes := binary.PutUvarint(tmp, uint64(posting-previousPosting))
		buf.Write(tmp[:numBytes])
		previousPosting = posting
	}

	return buf.Bytes()
}

func newTrieNode() *trieNode {
	node := new(trieNode)
	node.edges = make(map[rune]*trieNode, 0)
	node.postings = make(map[int]int, 0)
	return node
}

func NewTrieBuilder() TrieBuilder {
	root := newTrieNode()
	return root
}
