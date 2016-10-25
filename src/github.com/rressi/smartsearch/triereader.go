package smartsearch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// It represents a trie edge as it have been decoded by an instance of
// TrieReader
type Edge struct {
	Rune       rune // UNICODE code point associated with this edge
	nodeOffset int  // Byte offset of the target node of this edge
}

// It represents a trie node as it have been decoded by an instance of
// TrieReader.
//
// The content of this structure can be stored for later usage and given one
// binary index nodes are universally consistent. They can be marshaled and
// stored on files or passed between different services without problems.
type Node struct {
	NumPostings    int // Number of postings contained by the node
	NumEdges       int // Number of edges departing from this node
	postingsOffset int // Byte offset of the list of postings.
	edgesOffset    int // Byte offset of the list of edges.
}

// This error is returned when a passed offset is invalid.
var OutOfBounds = errors.New("Offset out of bound")

// A structure containing the state of a trie reader.
//
// It can be cloned in order to have two state machines decoding and traversing
// the trie from the current position. Clones are completely independent.
//
// These readers are working as state machines that decode lazily the bytes
// while traversing the trie. No up-front decoding of postings and edges is
// performed.
//
// They can jump to one state to another to save CPU resources (methods
// JumpNode, EnterNode).
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

// It creates a new TrieReader from the given bytes.
//
// Passed bytes are not changed by the created reader, just shared lazily
// decoded while traversing the trie. Changing this bytes may lead to undefined
// behaviour of the TrieReader.
//
// It returns:
// - the newly created TrieReader
// - information about its root node
// - an error in case of failure
func NewTrieReader(bytes_ []byte) (trieReader *TrieReader, node Node,
	err error) {
	trieReader = new(TrieReader)
	trieReader.bytes = bytes_
	node, err = trieReader.Reset()
	return
}

// It resets the state of this reader so that it restarts from the root node.
//
// It is equivalent to create a new reader with the very same bytes.
//
// It returns:
// - information about the root node.
// - an error in case of failure.
func (t *TrieReader) Reset() (_ Node, _ error) {
	t.reader = bytes.NewReader(t.bytes)
	return t.readNode()
}

// It returns the byte offset of the next byte to be decoded.
func (t *TrieReader) tell() (offset int) {
	offset = len(t.bytes) - t.reader.Len()
	return
}

// It sets the current TrieReader in a dummy and non-recoverable state after a
// critical failure.
//
// This happen only when some inconsistency is found into the input bytes.
func (t *TrieReader) clear() {
	t.reader = bytes.NewReader(t.bytes[:0])
	t.postingsLeft = 0
	t.edgesLeft = 0
	t.childrenBaseOffset = 0
	t.edgesOffset = 0
	t.posting = 0
	t.rune_ = 0
}

// It sets the offset of the next byte to be decoded.
//
// It returns:
// - an error in case of failure.
func (t *TrieReader) seek(offset int) (err error) {

	if int(offset) <= 0 || int(offset) > len(t.bytes) {
		err = OutOfBounds
	} else {
		t.reader = bytes.NewReader(t.bytes[offset:])
	}

	return
}

// It decodes one integer using Uvarint format.
//
// It returns:
// - the decoded value.
// - an error in case of failure.
func (t *TrieReader) readInt() (value int, err error) {
	var value_ uint64
	value_, err = binary.ReadUvarint(t.reader)
	value = int(value_)
	return
}

// It decodes the header of one node.
//
// It returns:
// - information about the decoded node.
// - an error in case of failure.
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

// It sets this reader to the given node.
//
// It can be used to go back to a node that have already been processed. The
// content of a node may come from a cache or some storage. It works only if
// have been generated by a binary identical trie.
//
// It returns:
// - an error in case of failure.
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

// It tests if there are no more postings to decode.
//
// In case it is true, it decodes the micro-header at the beginning of the
// list of edges.
//
// It returns:
// - an error in case of failure.
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

// It decodes next posting.
//
// When all postings have already been read it returns (0, io.EOF).
//
// Postings are decoded after entering one node and before decoding edges.
// An early call of method ReadEdge inhibits this method.
//
// It returns:
// - the decoded posting.
// - an error in case of failure.
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

// It decodes all remaining postings.
//
// If all postings have already been read it returns (nil, io.EOF).
//
// Postings are decoded after entering one node and before decoding edges.
// An early call of method ReadEdge inhibits this method.
//
// It returns:
// - all decoded posting in a sorted deduplicated array.
// - an error in case of failure.
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

// It decodes all remaining postings plus all the postings from all sub-nodes
// recursively.
//
// If all postings and edges have already been read it returns (nil, io.EOF).
//
// This method effectively consumes all remaining postings and edges of current
// node, the only things to do after it are:
// - jumping to another Node (method JumpNode).
// - entering in one node from one Edge returned previously (method EnterNode).
// - reset the reader to the root node (method Reset).
//
// It returns:
// - all decoded posting in a sorted deduplicated array.
// - an error in case of failure.
func (t *TrieReader) ReadAllPostingsRecursive() (postings []int,
	err error) {

	if t.postingsLeft == 0 && t.edgesLeft == 0 {
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

// It skips decoding of all the remaining postings preparing the reader to
// decode the edges.
//
// It returns:
// - an error in case of failure.
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

// It decodes one edge.
//
// An early call of this method inhibits reading of postings from the current
// node.
//
// Returned value can be used to enter the sub-node targeted by current edge.
//
// It returns:
// - information about the decoded edge.
// - an error in case of failure.
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

// It decodes all remaining edges.
//
// An early call of this method inhibits reading of postings from the current
// node.
//
// Returned values can be used to enter the sub-node targeted by current edge.
//
// It returns:
// - information about the decoded edges in a single array, sorted by relative
//   UNICODE code point.
// - an error in case of failure.
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

// It uses edge information returned by ReadEdge, ReadAllEdges to enter targeted
// child node.
//
// It returns:
// - information about the entered node.
// - an error in case of failure.
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

// Takes one term and traverses the trie from the current node matching all
// term's UNICODE code points in order.
//
// If during the traversal is not able to find one edge it returns a
// zeroed-node, this is not considered a failure but simply a negative match.
//
// It returns:
// - information about the final node.
// - an error in case of failure.
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
