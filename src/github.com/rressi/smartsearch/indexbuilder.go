package smartsearch

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
)

// IndexBuilder is a component that collects documents to generate one index
// that it allow to dump in a space-efficient blob that component Index is
// able to use for search.
type IndexBuilder interface {

	// It indexes a document, given an unique id and its content.
	//
	// If the same id is used many times it consider the passed content as part
	// of the same document.
	AddDocument(id int, content string)

	// It indexes a JSON document.
	//
	// Parameters:
	// - jsonDocument:  JSON document's bytes.
	// - idField:       JSON attribute for the unique id.
	// - contentFields: JSON attributes for the content to be indexed.
	//
	// Return:
	// - id:  The document id as it have extracted from the JSON document.
	// - err: An error in case of failure (in such case id is zero).
	//
	// Notes:
	// - the root object must be a dictionary
	// - it can access only to values of the root object.
	// - the unique id must be a positive integer, it is OK if it have been
	//   encoded as a string.
	// - if the same id is used many times it consider the passed content as
	//   part of the same document.
	AddJsonDocument(jsonDocument []byte, idField string,
		contentFields []string)

	// Applies method AddJsonDocument on all lines read from the passed
	// io.Reader.
	//
	// Return:
	// - numLines: Number of lines parsed.
	// - err:      An error in case of failure.
	//
	// Notes:
	// - it stops at the first failure.
	// - if the same id is used many times it consider the passed content as
	//   part of the same document.
	IndexJsonStream(reader io.Reader, idField string, contentFields []string) (
		numLines int, err error)

	// Applies method AddJsonDocument on all lines read from the passed
	// io.Reader and also returns a map mapping the document id with its raw
	// content.
	//
	// Return:
	// - documents: A map id -> JSON bytes from the processed documents
	// - err:       An error in case of failure (in such case documents is nil).
	//
	// Notes:
	// - it stops at the first failure.
	// - if the same id is used many times it fails.
	LoadAndIndexJsonStream(reader io.Reader, idField string,
		contentFields []string) (documents JsonDocuments, err error)

	// Generates a blob from all indexed documents and writes it to the passed
	// io.Writer.
	Dump(writer io.Writer) error

	// Aborts all pending co-routines, their job will be lost.
	Abort()
}

// Creates a new IndexBuilder.
//
// Warning: at first added content some go-routines are created to process the
// data concurrently. This go-routines are joined when methods
// IndexBuilder.Abort or IndexBuilder.Dump are called. To avoid leakages please
// use a deferred call to one of the 2 just after creating the builder.
func NewIndexBuilder() IndexBuilder {

	b := new(indexBuilderImpl)

	// Starts all the indexers:
	n := runtime.NumCPU()
	if n > 1 {
		for i := 0; i < n; i++ {
			b.indexers = append(b.indexers, NewConcurrentIndexer())
		}
	} else {
		b.indexers = append(b.indexers, NewIndexer())
	}

	return b
}

// Used to implement an IndexBuilder.
type indexBuilderImpl struct {
	indexers      []Indexer
	documentCount int
	trieBuilder   TrieBuilder
}

// Implementation of IndexBuilder.AddDocument
func (b *indexBuilderImpl) AddDocument(id int, content string) {
	k := b.documentCount % len(b.indexers)
	b.indexers[k].AddContent(id, []byte(content))
	b.documentCount++
}

// Implementation of IndexBuilder.AddJsonDocument
func (b *indexBuilderImpl) AddJsonDocument(jsonDocument []byte, idField string,
	contentFields []string) {
	k := b.documentCount % len(b.indexers)
	b.indexers[k].AddRawContent(jsonDocument,
		MakeJsonExtractor(idField, contentFields))
	b.documentCount++
	return
}

// Implementation of IndexBuilder.IndexJsonStream
func (b *indexBuilderImpl) IndexJsonStream(reader io.Reader, idField string,
	contentFields []string) (numLines int, err error) {

	// Any further failure will reset our state machine:
	defer func() {
		if err == io.EOF {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("IndexBuilder.IndexJsonStream, at line %d: %v",
				numLines, err)
		}
	}()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue // Ignores empty lines.
		}

		numLines += 1
		b.AddJsonDocument(scanner.Bytes(), idField, contentFields)
	}
	err = scanner.Err()

	return
}

// A map document id -> JSON bytes.
type JsonDocuments map[int][]byte

// Implementation of IndexBuilder.LoadAndIndexJsonStream
func (b *indexBuilderImpl) LoadAndIndexJsonStream(
	reader io.Reader,
	idField string,
	contentFields []string) (documents JsonDocuments, err error) {

	// Any further failure will reset our state machine:
	numLines := 0
	defer func() {
		if err == io.EOF {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("IndexBuilder.LoadAndIndexJsonStream, at line "+
				"%d: %v", numLines, err)
		}
	}()

	extractor := MakeJsonExtractor(idField, contentFields)
	documents_ := make(map[int][]byte, 0)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue // Ignores empty lines.
		}
		numLines += 1

		var id int
		var content string
		id, content, err = extractor(scanner.Bytes())
		if err != nil {
			return
		}

		if _, ok := documents_[id]; ok {
			err = fmt.Errorf("Duplicated document id %v", id)
			return
		}

		var blob []byte
		blob = make([]byte, len(scanner.Bytes()))
		copy(blob, scanner.Bytes())
		documents_[id] = blob

		b.AddDocument(id, content)
	}

	documents = documents_
	return
}

// Implementation of IndexBuilder.Dump
func (b *indexBuilderImpl) Dump(writer io.Writer) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("IndexBuilder.Dump: %v", err)
		}
	}()

	// We need a trie builder if not already built:
	if b.trieBuilder == nil {
		b.trieBuilder = NewConcurrentTrieBuilder()
	}

	// If there is pending content takes it from the indexers:
	var errors []error
	if len(b.indexers) > 0 {

		// Tells all the indexers to finish their job:
		for i := range b.indexers {
			b.indexers[i].Finish()
		}

		// Collects terms from the indexers:
		for i := range b.indexers {
			results, errors_ := b.indexers[i].Result()
			b.trieBuilder.AddBulk(results)
			if len(errors_) > 0 {
				errors = append(errors, errors_...)
			}
		}

		b.indexers = nil // They are useless now.
	}

	numErrors := len(errors)
	if numErrors > 0 {
		if numErrors == 1 {
			err = fmt.Errorf("IndexBuilder.Dump: %v", errors[0])
		} else {
			err = fmt.Errorf("IndexBuilder.Dump: %v errors occurred while "+
				"indexing", numErrors)
		}
		return
	}

	// Generates our blob:
	err = b.trieBuilder.Dump(writer)
	return
}

// Implementation of IndexBuilder.Abort
func (b *indexBuilderImpl) Abort() {

	// Stops all indexers:
	if len(b.indexers) > 0 {

		// Tells all the indexers to finish their job:
		for i := range b.indexers {
			b.indexers[i].Finish()
		}

		b.indexers = nil // They are useless now.
	}
}
