package smartsearch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

type IndexBuilder interface {
	AddDocument(id int, content string) error
	AddJsonDocument(jsonDocument []byte, idField string,
		contentFields []string) (id int, err error)
	IndexJsonStream(reader io.Reader, idField string, contentFields []string) (
		numLines int, err error)
	LoadAndIndexJsonStream(reader io.Reader, idField string,
		contentFields []string) (
		documents JsonDocuments, err error)
	Dump(writer io.Writer) error
}

func NewIndexBuilder() IndexBuilder {
	b := new(indexBuilderImpl)
	b.trie = NewTrieBuilder()
	return b
}

type indexBuilderImpl struct {
	trie TrieBuilder
}

func (b *indexBuilderImpl) AddDocument(id int, content string) (_ error) {
	for _, term := range Tokenize(content) {
		b.trie.Add(id, term)
	}

	return
}

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

type JsonDocuments map[int][]byte

func (b *indexBuilderImpl) LoadAndIndexJsonStream(
	reader io.Reader, idField string, contentFields []string) (
	documents JsonDocuments, err error) {

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
		if _, ok := documents[id]; ok {
			err = fmt.Errorf("Duplicated document id %v", id)
			return
		}

		// Indexes current document:
		documents_[id] = scanner.Bytes()
	}

	documents = documents_
	return
}

func (b *indexBuilderImpl) Dump(writer io.Writer) error {
	return b.trie.Dump(writer)
}
