package smartsearch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sort"
	"unicode/utf8"
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

	// It adds many terms that have been already nicely indexed.
	//
	// Data is passed as a sorted list of terms, each term is packed together
	// with its postings (sorted and deduplicated) and the number of
	// times this term have been found.
	//
	// If a term contains an empty string then the posting is added to the root
	// node.
	AddBulk(data IndexedTerms)

	// Generates a trie and serializes to the passed io.Writer.
	//
	// It returns error on failures.
	Dump(dst io.Writer) error
}

// Creates a new TrieBuilder.
func NewTrieBuilder() TrieBuilder {
	return newTrieNode()
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
			node = node.enterChild(rune_)
		}
	}
	node.postings = append(node.postings, posting)
	node.appendedPostings += 1
}

// Implementation of TrieBuilder.AddBulk
func (t *trieNode) AddBulk(data IndexedTerms) {

	// We need to know how many runes we need to memoize while importing the
	// data:
	requiredRunes := 0 // one for the root node
	for _, indexedTerm := range data {
		if requiredRunes < len(indexedTerm.term) {
			requiredRunes = len(indexedTerm.term)
		}
	}
	requiredRunes++

	// We need a stack of nodes and a stack of runes in order to be able to
	// take advantage of the common prefixes that sorted terms have naturally:
	nodes := make([]*trieNode, requiredRunes)
	nodes[0] = t
	runes := make([]rune, requiredRunes)
	runes[0] = 0
	currPosition := 0

	// For each term in a sorted order:
	for _, indexedTerm := range data {

		if len(indexedTerm.postings) == 0 {
			continue
		}

		// Walks to the target node starting from the last node sharing the
		// same prefix with last one previous:
		node := t
		for i, rune_ := range indexedTerm.term {
			j := i + 1
			if j <= currPosition && runes[j] == rune_ {
				node = nodes[j] // Prefix match
			} else {
				currPosition = j // Prefix match stops here
				runes[j] = rune_
				node = nodes[i].enterChild(rune_)
				nodes[j] = node
			}
		}

		if len(t.postings) > 0 && node.appendedPostings > 0 {
			node.postings = SortDedupPostings(node.postings)
			node.appendedPostings = 0
		}
		node.postings = UnitePostings(node.postings, indexedTerm.postings)
	}
}

// It implements TrieBuilder.Dump
func (t *trieNode) Dump(dst io.Writer) error {
	_, err := t.dumpRec(dst)
	if err != nil {
		err = fmt.Errorf("TrieBuilder.Dump: %v", err)
	}
	return err
}

// Returns child node given the related rune, if needed creates it.
func (t *trieNode) enterChild(r rune) (childNode *trieNode) {
	var ok bool
	childNode, ok = t.edges[r]
	if !ok {
		childNode = newTrieNode()
		t.edges[r] = childNode
	}
	return
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

type trieConcurrentSlave struct {
	inChan  chan interface{}
	outChan chan bool
}

type trieConcurrentAdd struct {
	node    *trieNode
	term    string
	posting int
}

type trieConcurrentAddBulk struct {
	node     *trieNode
	term     string
	postings []int
}

func newTrieConcurrentSlave() *trieConcurrentSlave {
	slave := new(trieConcurrentSlave)
	slave.inChan = make(chan interface{}, 100)
	slave.outChan = make(chan bool, 1)
	return slave
}

func (s *trieConcurrentSlave) Loop() {

	for command := range s.inChan {
		switch data := command.(type) {
		case trieConcurrentAdd:
			data.node.Add(data.posting, data.term)
		case trieConcurrentAddBulk:
			data.node.AddBulk(IndexedTerms{{
				term:     data.term,
				postings: data.postings}})
		case bool:
			s.outChan <- true
			return
		}
	}
}

// -----------------------------------------------------------------------------

type trieConcurrentMaster struct {
	root   *trieNode
	slaves []*trieConcurrentSlave
}

// Creates a new TrieBuilder that works concurrently:
func NewConcurrentTrieBuilder() TrieBuilder {

	numCores := runtime.NumCPU()
	if numCores <= 1 {
		return newTrieNode()
	}

	master := new(trieConcurrentMaster)
	master.root = newTrieNode()

	master.slaves = make([]*trieConcurrentSlave, numCores)
	for i := 0; i < numCores; i++ {
		slave := newTrieConcurrentSlave()
		go slave.Loop()
		master.slaves[i] = slave
	}

	return master
}

func (m *trieConcurrentMaster) Add(posting int, term string) {

	// Void term goes to the root node that is not assigned to any slave:
	if len(term) == 0 {
		m.root.Add(posting, term)
		return
	}

	// Takes the first rune and uses it to select one slave for this job;
	rune_, sz := utf8.DecodeRuneInString(term)
	k := int(rune_) % len(m.slaves)

	// Passes this command to the slave:
	m.slaves[k].inChan <- trieConcurrentAdd{
		node:    m.root.enterChild(rune_),
		posting: posting,
		term:    term[sz:]}
}

func (m *trieConcurrentMaster) AddBulk(data IndexedTerms) {

	for _, datum := range data {

		if len(datum.postings) == 0 {
			continue
		}

		// Void term goes to the root node that is not assigned to any slave:
		if len(datum.term) == 0 {
			m.root.AddBulk(IndexedTerms{datum})
			continue
		}

		rune_, sz := utf8.DecodeRuneInString(datum.term)
		k := int(rune_) % len(m.slaves)

		m.slaves[k].inChan <- trieConcurrentAddBulk{
			node:     m.root.enterChild(rune_),
			postings: datum.postings,
			term:     datum.term[sz:]}
	}
}

func (m *trieConcurrentMaster) Dump(dst io.Writer) error {

	// Signals all slave to terminate:
	for _, slave := range m.slaves {
		slave.inChan <- true
	}

	// Joins all slaves:
	for _, slave := range m.slaves {
		<-slave.outChan
	}

	return m.root.Dump(dst)
}
