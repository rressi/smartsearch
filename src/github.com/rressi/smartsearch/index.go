package smartsearch

import (
	"bytes"
	"fmt"
	"io"
)

// An interface to search using pre-build indices.
type Index interface {

	// Given the passed query, it searches it inside the index and returns all the
	// postings of matching documents.
	//
	// It returns:
	// - postings of matching documents, sorted and deduplicated.
	// - an error in case of failure
	Search(query string, limit int) (postings []int, err error)
}

// Given the passed io.Reader, loads an index previously generated with
// IndexBuilder.
//
// It returns:
// - the newly created index.
// - the bytes containing the read index.
// - an error on failure.
func NewIndex(reader io.Reader) (index Index, rawIdex []byte, err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("NewIndex:: %v", err)
		}
	}()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return
	}

	index_ := new(indexImpl)
	index_.trie, _, err = NewTrieReader(buf.Bytes())
	if err != nil {
		return
	}

	index = index_
	rawIdex = buf.Bytes()
	return
}

// Local storage for the private implementation of an Index.
type indexImpl struct {
	trie      *TrieReader
	tokenizer Tokenizer
}

// Private implementation of Index.Search.
func (idx *indexImpl) Search(query string, limit int) (
	postings []int, err error) {

	defer func() {
		if err == io.EOF {
			// Simply there were nor results from one term:
			err = nil
		}
		if err != nil {
			err = fmt.Errorf("Index.Search '%v': %v", query, err)
		}
	}()

	if limit == 0 {
		return // Nothing to do.
	}

	if idx.tokenizer == nil {
		idx.tokenizer = NewTokenizer()
	}

	// Extracts all the terms:
	terms, incomplete_term, err := idx.tokenizer.ForSearch(query)
	if err != nil {
		return
	}

	// Special case: we need to extract all the postings:
	if len(terms) == 0 && len(incomplete_term) == 0 {
		idx.trie.Reset()
		var _postings []int
		_postings, err = idx.trie.ReadAllPostingsRecursive()
		if err != nil {
			return
		}

		postings = _postings
		return
	}

	// Performs exact match with all complete terms while intersecting all
	// the fetched postings:
	var mergedPostings []int
	for i, term := range terms {
		var node Node
		idx.trie.Reset()
		node, err = idx.trie.Match(term)
		if err != nil || node.NumPostings == 0 {
			return
		}

		var nodePostings []int
		nodePostings, err = idx.trie.ReadAllPostings()
		if err != nil {
			return
		}

		if i == 0 {
			mergedPostings = nodePostings
		} else {
			mergedPostings = IntersectPostings(mergedPostings, nodePostings)
			if len(mergedPostings) == 0 {
				return // No result!
			}
		}
	}

	// If there was a potentially incomplete term, it performs prefix match with
	// it and intersects all resulting postings with ones previously obtained:
	if len(incomplete_term) > 0 {

		var node Node
		idx.trie.Reset()
		node, err = idx.trie.Match(incomplete_term)
		if err != nil ||
			(node.NumPostings == 0 && node.NumEdges == 0) {
			return
		}

		var nodePostings []int
		nodePostings, err = idx.trie.ReadAllPostingsRecursive()
		if err != nil {
			return
		}

		if mergedPostings == nil {
			mergedPostings = nodePostings
		} else {
			mergedPostings = IntersectPostings(mergedPostings, nodePostings)
			if len(mergedPostings) == 0 {
				return // No result!
			}
		}
	}

	// In case we have a limit set it truncates the result:
	if limit >= 0 && limit < len(mergedPostings) {
		postings = mergedPostings[:limit]
	} else {
		postings = mergedPostings
	}

	return
}
