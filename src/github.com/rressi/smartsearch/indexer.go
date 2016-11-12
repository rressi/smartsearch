package smartsearch

import (
	"sort"
)

// A component to tokenize documents with a slave co-routine
type Indexer interface {

	// Posts new content to be processed.
	AddDocument(id int, content string)

	// Terminates the slave go-routine.
	Finish()

	// Wait for termination and fetches the final result.
	Result() IndexedTerms
}

// Used by implementation of Indexer to receive new input.
type indexerInput struct {
	id      int
	content string
}

// Main struct used by implementation of Indexer.
type indexerImpl struct {
	terms     map[string][]int
	tokenizer Tokenizer
	inChan    chan<- indexerInput
	outChan   <-chan IndexedTerms
}

// Implementation of IndexTokenizer.AddDocument
func (i *indexerImpl) AddDocument(id int, content string) {
	command := indexerInput{id: id, content: content}
	i.inChan <- command
}

// Implementation of IndexTokenizer.Done
func (i *indexerImpl) Finish() {
	i.inChan <- indexerInput{id: -1, content: ""}
}

// Implementation of IndexTokenizer.Result
func (i *indexerImpl) Result() IndexedTerms {
	return <-i.outChan
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
			if command.id >= 0 {
				terms := i.tokenizer.Apply(command.content)
				for _, term := range terms {
					i.terms[term] = append(i.terms[term], command.id)
				}
			} else {
				// Generates the final result and returns.
				var results IndexedTerms
				for term, postings := range i.terms {
					result := IndexedTerm{
						term:        term,
						postings:    SortDedupPostings(postings),
						occurrences: len(postings)}
					results = append(results, result)
				}
				sort.Sort(results)
				outChan <- results
				return // End of story.
			}
		}
	}() // go func
	i.inChan = inChan
	i.outChan = outChan
	return i
}
