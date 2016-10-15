package smartsearch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

type Edge struct {
	Rune       rune
	nodeOffset int
}

type Node struct {
	NumPostings    int
	NumEdges       int
	postingsOffset int
	edgesOffset    int
}

var OutOfBounds = errors.New("Out of bound")
var EOF = errors.New("Invalid mode")

type TrieReader struct {
	bytes              []byte
	reader             *bytes.Reader
	postingsLeft       int
	edgesLeft          int
	edgesOffset        int
	childrenBaseOffset int
	posting            int
}

func NewFromMemory(bytes_ []byte) (trieReader *TrieReader, node Node,
	err error) {
	trieReader = new(TrieReader)
	trieReader.bytes = bytes_
	trieReader.reader = new(bytes.Reader)
	node, err = trieReader.Reset()
	return
}

func (t *TrieReader) Reset() (_ Node, _ error) {
	t.reader.Reset(t.bytes)
	return t.readNode()
}

func (t *TrieReader) tell() (offset int) {
	offset = len(t.bytes) - t.reader.Len()
	return
}

func (t *TrieReader) clear() {
	t.reader.Reset(t.bytes[:0])
	t.postingsLeft = 0
	t.edgesLeft = 0
	t.childrenBaseOffset = 0
	t.edgesOffset = 0
	t.posting = 0
}

func (trieReader *TrieReader) seek(offset int) (err error) {

	if int(offset) <= 0 || int(offset) > len(trieReader.bytes) {
		err = OutOfBounds
	} else {
		trieReader.reader.Reset(trieReader.bytes[offset:])
	}

	return
}

func (t *TrieReader) readInt() (value int, err error) {
	var value_ uint64
	value_, err = binary.ReadUvarint(t.reader)
	value = int(value_)
	return
}

func (t *TrieReader) readNode() (node Node, err error) {

	if t.reader.Len() == 0 {
		err = EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.readNode: %v", err)
		}
	}()

	t.postingsLeft, err = t.readInt()
	if err != nil {
		return
	}

	t.edgesLeft, err = t.readInt()
	if err != nil {
		return
	}

	var sizeOfPosting int
	sizeOfPosting, err = t.readInt()
	if err != nil {
		return
	}

	postingsOffset := t.tell()
	t.edgesOffset = postingsOffset + sizeOfPosting
	t.childrenBaseOffset = 0
	t.posting = 0

	node = Node{t.postingsLeft, t.edgesLeft, postingsOffset, t.edgesOffset}
	return
}

func (t *TrieReader) testEndPostings() (err error) {
	if t.postingsLeft == 0 {

		var sizeOfEdges int
		sizeOfEdges, err = t.readInt()
		if err != nil {
			err = fmt.Errorf("TrieReader.testEndPostings: %v", err)
			return
		}

		t.postingsLeft = 0
		t.edgesOffset = 0
		t.childrenBaseOffset = t.tell() + sizeOfEdges
	}
	return
}

func (t *TrieReader) ReadPosting() (posting int, err error) {

	if t.postingsLeft == 0 {
		err = EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("ReadPosting: %v", err)
		}
	}()

	var increment = 0
	increment, err = t.readInt()
	if err != nil {
		return
	}

	t.posting += int(increment)
	t.postingsLeft--
	posting = t.posting

	t.testEndPostings()
	return
}

func (t *TrieReader) ReadAllPostings() (postings []int, err error) {

	postings = make([]int, t.postingsLeft)
	var i int
	for t.postingsLeft > 0 {
		postings[i], err = t.ReadPosting()
		if err != nil {
			break
		}
		i++
		t.postingsLeft--
	}

	return
}

func (t *TrieReader) ReadAllPostingsRecursive() (postings []int,
	err error) {

	if t.postingsLeft == 0 {
		err = EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil && err != EOF {
			t.clear()
			err = fmt.Errorf("TrieReader.ReadAllPostingsRecursive: %v", err)
		}
	}()

	postings = make([]int, 0)

	// Prepares a queue for a breath-first traversal of the trie:
	workingQueue := make([]*TrieReader, 1)
	workingQueue[0] = t

	// Main loop:
	for len(workingQueue) > 0 {
		currentReader := workingQueue[0]
		workingQueue = workingQueue[1:]

		// Reads all the postings and appends them to the result:
		for currentReader.postingsLeft > 0 {
			var posting int
			posting, err = currentReader.ReadPosting()
			if err != nil {
				return
			}
			postings = append(postings, posting)
		}

		// Iterates through all the edges and queue spawn readers to iterate
		// them later:
		for currentReader.edgesLeft > 0 {
			var edge Edge
			edge, err = currentReader.ReadEdge()
			if err != nil {
				return
			}

			// Spawn the current trie and jumps it to the edge's target node
			childTrie := new(TrieReader)
			*childTrie = *currentReader
			_, err = childTrie.EnterNode(edge)
			if err != nil {
				return
			}

			workingQueue = append(workingQueue, childTrie)
		}
	}

	// Sorts and deduplicates all collected postings:
	if len(postings) > 1 {
		sort.Ints(postings)
		var nextUnique int
		for i := 1; i < len(postings); i++ {
			if postings[nextUnique] != postings[i] {
				nextUnique++
				postings[nextUnique] = postings[i]
			}
		}
		postings = postings[:nextUnique+1]
	}

	return
}

func (t *TrieReader) SkipPostings() (err error) {

	if t.edgesOffset == 0 {
		err = EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.SkipPostings: %v", err)
		}
	}()

	err = t.seek(t.edgesOffset)
	if err != nil {
		return
	}

	t.postingsLeft = 0
	t.testEndPostings()
	return
}

func (t *TrieReader) ReadEdge() (edge Edge, err error) {

	if t.edgesLeft == 0 {
		err = EOF
		return
	}

	// Any further failure will reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.ReadEdge: %v", err)
		}
	}()

	if t.postingsLeft > 0 {
		err = t.SkipPostings()
		if err != nil {
			return
		}
	}

	var rune_ int
	rune_, err = t.readInt()
	if err != nil {
		return
	}

	var sizeOfChildrenNode int
	sizeOfChildrenNode, err = t.readInt()
	if err != nil {
		return
	}

	edge = Edge{rune(rune_), t.childrenBaseOffset}
	t.childrenBaseOffset += sizeOfChildrenNode
	t.edgesLeft--
	if t.edgesLeft == 0 {
		err = EOF
	}
	return
}

func (trieReader *TrieReader) ReadAllEdges() (edges []Edge, err error) {
	edges = make([]Edge, trieReader.edgesLeft)
	var i int
	for err == nil && trieReader.edgesLeft > 0 {
		edges[i], err = trieReader.ReadEdge()
		i++
	}
	if err != nil && err != EOF {
		err = fmt.Errorf("TrieReader.ReadAllEdges: %v", err)
	}
	return
}

func (source *TrieReader) EnterNode(edge Edge) (node Node, err error) {

	err = source.seek(edge.nodeOffset)
	if err != nil {
		err = fmt.Errorf("TrieReader.EnterNode: %v", err)
		return
	}

	node, err = source.readNode()
	return
}
