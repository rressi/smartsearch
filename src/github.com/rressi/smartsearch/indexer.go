package smartsearch

import (
	"fmt"
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
	Result() (result IndexedTerms, errors []error)
}

// Used by implementation of Indexer to receive new input.
type indexerInput struct {
	id        int // negative values are used to control the flow
	content   []byte
	extractor ContentExtractor
}

// Main struct used by implementation of Indexer.
type indexerImpl struct {
	terms     map[string][]int
	tokenizer Tokenizer
	inChan    chan<- indexerInput
	outChan   <-chan bool
	results   IndexedTerms
	errors    []error
}

// Implementation of IndexTokenizer.AddDocument
func (i *indexerImpl) AddContent(id int, content []byte) {
	if len(content) == 0 {
		// pass
	} else if id < 0 {
		err := fmt.Errorf("indexerImpl.AddContent: invalid id %v ", id)
		i.errors = append(i.errors, err)
	} else if i.inChan != nil {
		command := indexerInput{id: id, content: content}
		i.inChan <- command
	} else {
		i.onAddContent(id, string(content))
	}
}

// Indexes the content
func (i *indexerImpl) onAddContent(id int, content string) {
	terms := i.tokenizer.Apply(content)
	for _, term := range terms {
		i.terms[term] = append(i.terms[term], id)
	}
}

// Posts new content to be processed.
func (i *indexerImpl) AddRawContent(raw []byte, extractor ContentExtractor) {
	if len(raw) == 0 {
		// pass
	} else if i.inChan != nil {
		command := indexerInput{content: raw, extractor: extractor}
		i.inChan <- command
	} else {
		id, content := i.onAddRawContent(raw, extractor)
		if len(content) >= 0 {
			i.onAddContent(id, content)
		}
	}
}

func (i *indexerImpl) onAddRawContent(raw []byte, extractor ContentExtractor) (
	id int, content string) {

	// Uses the extractor to get the pure content and the document id.
	id, content, err := extractor(raw)
	if err != nil {
		err = fmt.Errorf("indexerImpl.onAddRawContent: %v", err)
	} else if id < 0 {
		err = fmt.Errorf("indexerImpl.onAddRawContent: invalid id %v "+
			"extracted", id)
	}
	if err != nil {
		i.errors = append(i.errors, err)
		id = -1
		content = ""
	}

	return
}

// Implementation of IndexTokenizer.Finish
func (i *indexerImpl) Finish() {
	if i.inChan != nil {
		i.inChan <- indexerInput{id: -1}
	} else {
		i.onFinish()
	}
}

// Generates the final result:
func (i *indexerImpl) onFinish() {
	for term, postings := range i.terms {
		result := IndexedTerm{
			term:        term,
			postings:    SortDedupPostings(postings),
			occurrences: len(postings)}
		i.results = append(i.results, result)
	}
	sort.Sort(i.results)
	return
}

// Implementation of IndexTokenizer.Result
func (i *indexerImpl) Result() (results IndexedTerms, errors []error) {
	if i.outChan != nil {
		<-i.outChan // Waits for termination.
	}
	errors = i.errors
	results = i.results
	return
}

// Creates an IndexTokenizer.
func NewIndexer() Indexer {
	i := new(indexerImpl)
	i.tokenizer = NewTokenizer()
	i.terms = make(map[string][]int)
	return i
}

// Creates an IndexTokenizer that uses an internal go-routine for the heavy job.
func NewConcurrentIndexer() Indexer {
	inChan := make(chan indexerInput, 1000)
	outChan := make(chan bool, 1)
	i := new(indexerImpl)
	i.tokenizer = NewTokenizer()
	i.terms = make(map[string][]int)
	go i.concurrentWorker(inChan, outChan)
	i.inChan = inChan
	i.outChan = outChan
	return i
}

func (i *indexerImpl) concurrentWorker(
	inChan chan indexerInput,
	outChan chan bool) {

	for command := range inChan {
		if command.extractor != nil {
			id, content := i.onAddRawContent(command.content, command.extractor)
			if len(content) > 0 {
				i.onAddContent(id, content)
			}
		} else if command.id >= 0 {
			i.onAddContent(command.id, string(command.content))
		} else {
			i.onFinish()
			outChan <- true
			break // Done!
		}
	}
}
