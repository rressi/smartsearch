package smartsearch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// IndexBuilder is a component that collects documents to generate one index
// that it allow to dump in a space-efficient blob that component Index is
// able to use for search.
type IndexBuilder interface {

	// It indexes a document, given an unique id and its content.
	//
	// If the same id is used many times it consider the passed content as part
	// of the same document.
	AddDocument(id int, content string) error

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
		contentFields []string) (id int, err error)

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
}

// Creates a new IndexBuilder.
func NewIndexBuilder() IndexBuilder {
	b := new(indexBuilderImpl)
	b.trie = NewTrieBuilder()
	return b
}

// Used to implement and IndexBuilder.
type indexBuilderImpl struct {
	trie TrieBuilder
}

// Implementation of IndexBuilder.AddDocument
func (b *indexBuilderImpl) AddDocument(id int, content string) (_ error) {
	for _, term := range Tokenize(content) {
		b.trie.Add(id, term)
	}

	return
}

// Implementation of IndexBuilder.AddJsonDocument
func (b *indexBuilderImpl) AddJsonDocument(jsonDocument []byte, idField string,
	contentFields []string) (id int, err error) {

	// Any further failure will reset our state machine:
	defer func() {
		if err == io.EOF {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("IndexBuilder.AddJsonDocument: %v", err)
		}
	}()

	var datum map[string]interface{}
	err = json.Unmarshal(jsonDocument, &datum)
	if err != nil {
		return
	}

	idRaw, ok := datum[idField]
	if !ok {
		err = fmt.Errorf("document does not have ID field '%v' defined",
			idField)
		return
	}

	var id_ int
	id_, err = strconv.Atoi(fmt.Sprint(idRaw))
	if err != nil {
		return
	}

	for _, field := range contentFields {
		content_, ok := datum[field]
		if ok {
			content := fmt.Sprint(content_)
			err = b.AddDocument(id_, content)
			if err != nil {
				return
			}
		}
	}

	id = id_
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
		_, err = b.AddJsonDocument(scanner.Bytes(), idField, contentFields)
		if err != nil {
			return
		}
	}

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

	documents_ := make(map[int][]byte, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue // Ignores empty lines.
		}
		numLines += 1

		var id int
		id, err = b.AddJsonDocument(scanner.Bytes(), idField, contentFields)
		if err != nil {
			return
		}

		if _, ok := documents_[id]; ok {
			err = fmt.Errorf("Duplicated document id %v", id)
			return
		}

		// Indexes current document:
		documents_[id] = scanner.Bytes()
	}

	documents = documents_
	return
}

// Implementation of IndexBuilder.Dump
func (b *indexBuilderImpl) Dump(writer io.Writer) error {
	return b.trie.Dump(writer)
}
