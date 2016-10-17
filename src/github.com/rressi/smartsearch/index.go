package smartsearch

import (
	"bytes"
	"fmt"
	"io"
)

type Index interface {
	Search(query string) (postings []int, err error)
}

func NewIndex(reader io.Reader) (index Index, err error) {

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
	return
}

type indexImpl struct {
	trie *TrieReader
}

func (idx *indexImpl) Search(query string) (postings []int, err error) {

	defer func() {
		if err == io.EOF {
			// Simply there were nor results from one term:
			err = nil
		}
		if err != nil {
			err = fmt.Errorf("Index.Search '%v': %v", query, err)
		}
	}()

	terms := Tokenize(query)
	if len(terms) == 0 {
		return
	}

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
			mergedPostings = MergePostings(mergedPostings, nodePostings)
			if len(mergedPostings) == 0 {
				return // NO result!
			}
		}
	}

	postings = mergedPostings
	return
}
