package smartsearch

import (
	"fmt"
	"io"
	"sort"
)

// A component to tokenize documents with a slave co-routine
type Indexer interface {

	// Posts new content to be indexed.
	AddContent(id int, content []byte)

	// Posts raw bytes with content to be extracted and indexed.
	AddRawContent(raw []byte, extractor ContentExtractor)

	// Terminates the slave go-routine.
	Finish()

	// Wait for termination and fetches the final result.
	Result() (result IndexedTerms, err error)
}

// Used by implementation of Indexer to receive new input.
type indexerInput struct {
	id        int
	content   []byte
	extractor ContentExtractor
}

func (i *indexerInput) extract() (id int, content string, err error) {

	if i.extractor == nil {
		if i.id < 0 {
			err = io.EOF
		} else {
			id = i.id
			content = string(i.content)
		}
		return
	}

	id_, content_, err_ := i.extractor(i.content)
	if err_ != nil {
		err = fmt.Errorf("indexerInput.Extract: %v", err_)
		return
	} else if id_ < 0 {
		err = fmt.Errorf("indexerInput.Extract: invalid id %v extracted", id_)
		return
	}

	id = id_
	content = content_
	return
}

// Main struct used by implementation of Indexer.
type indexerImpl struct {
	terms     map[string][]int
	tokenizer Tokenizer
	inChan    chan<- indexerInput
	outChan   <-chan IndexedTerms
	err       error
}

// Implementation of IndexTokenizer.AddDocument
func (i *indexerImpl) AddContent(id int, content []byte) {
	command := indexerInput{id: id, content: content}
	i.inChan <- command
}

// Posts new content to be processed.
func (i *indexerImpl) AddRawContent(raw []byte, extractor ContentExtractor) {
	command := indexerInput{content: raw, extractor: extractor}
	i.inChan <- command
}

// Implementation of IndexTokenizer.Done
func (i *indexerImpl) Finish() {
	i.inChan <- indexerInput{id: -1}
}

// Implementation of IndexTokenizer.Result
func (i *indexerImpl) Result() (result IndexedTerms, err error) {
	if i.err != io.EOF && i.err != nil {
		err = i.err
	}
	result = <-i.outChan
	return
}

// Creates an IndexTokenizer
func NewIndexer() Indexer {
	i := new(indexerImpl)
	i.tokenizer = NewTokenizer()
	inChan := make(chan indexerInput, 1000)
	outChan := make(chan IndexedTerms, 1)
	go func() {
		i.terms = make(map[string][]int)
		for command := range inChan {
			id, content, err := command.extract()
			if err != nil {
				var results IndexedTerms
				if err == io.EOF {
					// Generates the final result:
					for term, postings := range i.terms {
						result := IndexedTerm{
							term:        term,
							postings:    SortDedupPostings(postings),
							occurrences: len(postings)}
						results = append(results, result)
					}
					sort.Sort(results)
				}
				outChan <- results
				return // End of story.
			} else {
				// Indexes the content:
				terms := i.tokenizer.Apply(content)
				for _, term := range terms {
					i.terms[term] = append(i.terms[term], id)
				}
			}
		}
	}() // go func
	i.inChan = inChan
	i.outChan = outChan
	return i
}
