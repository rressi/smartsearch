package smartsearch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

var OutOfBounds = errors.New("Offset out of bound")

type TrieReader struct {
	bytes              []byte
	reader             *bytes.Reader
	postingsLeft       int
	edgesLeft          int
	edgesOffset        int
	childrenBaseOffset int
	posting            int
	rune_              int
}

func NewTrieReader(bytes_ []byte) (trieReader *TrieReader, node Node,
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
	t.rune_ = 0
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
		err = io.EOF
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
	if t.postingsLeft > 0 {
		sizeOfPosting, err = t.readInt()
		if err != nil {
			return
		}
	}

	postingsOffset := t.tell()
	t.edgesOffset = postingsOffset + sizeOfPosting
	t.childrenBaseOffset = 0
	t.posting = 0
	t.rune_ = 0

	// When there are no postings we need to prepare our machine to read nodes:
	err = t.testEndPostings()

	node = Node{t.postingsLeft, t.edgesLeft, postingsOffset, t.edgesOffset}
	return
}

func (t *TrieReader) JumpNode(node Node) (err error) {

	err = t.seek(node.postingsOffset)
	if err != nil {
		t.clear()
		return
	}

	t.postingsLeft = node.NumPostings
	t.edgesLeft = node.NumEdges
	t.edgesOffset = node.edgesOffset
	t.childrenBaseOffset = 0
	t.posting = 0
	t.rune_ = 0

	// When there are no postings we need to prepare our machine to read nodes:
	err = t.testEndPostings()

	return
}

func (t *TrieReader) testEndPostings() (err error) {

	// Just after reading the last posting, we need to prepare our state
	// machine to read edges:
	if t.postingsLeft == 0 && t.childrenBaseOffset == 0 && t.edgesLeft > 0 {

		var sizeOfEdges int
		sizeOfEdges, err = t.readInt()
		if err != nil {
			err = fmt.Errorf("TrieReader.testEndPostings: %v", err)
			t.clear()
			return
		}

		t.childrenBaseOffset = t.tell() + sizeOfEdges
	}

	return
}

func (t *TrieReader) ReadPosting() (posting int, err error) {

	if t.postingsLeft == 0 {
		err = io.EOF
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

	// When there are no postings we need to prepare our machine to read nodes:
	err = t.testEndPostings()

	posting = t.posting
	return
}

func (t *TrieReader) ReadAllPostings() (postings []int, err error) {

	if t.postingsLeft == 0 {
		err = io.EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.ReadAllPostings: %v", err)
		}
	}()

	num := t.postingsLeft
	postings_ := make([]int, t.postingsLeft)
	for i := 0; err == nil && i < num; i++ {
		postings_[i], err = t.ReadPosting()
	}

	if err == nil {
		postings = postings_
	}

	return
}

func (t *TrieReader) ReadAllPostingsRecursive() (postings []int,
	err error) {

	if t.postingsLeft == 0 {
		err = io.EOF
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil && err != io.EOF {
			t.clear()
			err = fmt.Errorf("TrieReader.ReadAllPostingsRecursive: %v", err)
		}
	}()

	postings_ := make([]int, 0)

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
			postings_ = append(postings_, posting)
		}

		// Iterates through all the edges and queue spawn readers to iterate
		// them later:
		for currentReader.edgesLeft > 0 {
			var edge Edge
			edge, err = currentReader.ReadEdge()
			if err != nil {
				return
			}

			// Spawn the current reader and jumps it to the edge's target node
			childTrie := new(TrieReader)
			*childTrie = *currentReader
			_, err = childTrie.EnterNode(edge)
			if err != nil {
				return
			}

			workingQueue = append(workingQueue, childTrie)
		}
	}

	postings = SortDedupPostings(postings_)
	return
}

func (t *TrieReader) skipPostings() (err error) {

	if t.postingsLeft == 0 {
		return
	}

	// Any further failure would reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.skipPostings: %v", err)
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
		err = io.EOF
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
		err = t.skipPostings()
		if err != nil {
			return
		}
	}

	var runeIncrement int
	runeIncrement, err = t.readInt()
	if err != nil {
		return
	}
	t.rune_ += runeIncrement

	var sizeOfChildrenNode int
	sizeOfChildrenNode, err = t.readInt()
	if err != nil {
		return
	}

	edge = Edge{rune(t.rune_), t.childrenBaseOffset}
	t.childrenBaseOffset += sizeOfChildrenNode
	t.edgesLeft--
	return
}

func (t *TrieReader) ReadAllEdges() (edges []Edge, err error) {

	if t.edgesLeft == 0 {
		err = io.EOF
		return
	}

	edges_ := make([]Edge, t.edgesLeft)
	var i int
	for err == nil && t.edgesLeft > 0 {
		edges_[i], err = t.ReadEdge()
		i++
	}
	if err != nil {
		err = fmt.Errorf("TrieReader.ReadAllEdges: %v", err)
		return
	}

	edges = edges_
	return
}

func (t *TrieReader) EnterNode(edge Edge) (node Node, err error) {

	// Any further failure will reset our state machine:
	defer func() {
		if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.EnterNode: %v", err)
		}
	}()

	err = t.seek(edge.nodeOffset)
	if err != nil {
		return
	}

	node, err = t.readNode()
	return
}

func (t *TrieReader) Match(term string) (node Node, err error) {

	// Handles post-condition:
	defer func() {
		if err == io.EOF {
			err = nil // No match have been found.
		} else if err != nil {
			t.clear()
			err = fmt.Errorf("TrieReader.Match('%v'): %v", term, err)
		}
	}()

	var node_ Node
	for _, targetRune := range term {
		var edge Edge
		for err == nil && edge.Rune < targetRune {
			edge, err = t.ReadEdge()
		}
		if err == nil && edge.Rune != targetRune {
			err = io.EOF // Not found!
		}
		if err != nil {
			return // Edge with the given rune have not found.
		}

		node_, err = t.EnterNode(edge)
		if err != nil {
			return // Edge with the given rune have not found.
		}
	}

	node = node_
	return
}
