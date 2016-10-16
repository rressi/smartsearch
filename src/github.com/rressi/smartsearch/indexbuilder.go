package smartsearch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type IndexBuilder interface {
	AddDocument(id int, content string) error
	ScanJsonStream(reader io.Reader, idField string,
		contentFields []string) (err error)
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

func (b *indexBuilderImpl) ScanJsonStream(reader io.Reader, idField string,
	contentFields []string) (err error) {

	// Any further failure will reset our state machine:
	defer func() {
		if err == io.EOF {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("IndexBuilder.AddJsonDocuments: %v", err)
		}
	}()

	scanner := bufio.NewScanner(reader)
	line := 0
	for scanner.Scan() {
		var datum map[string]interface{}
		err = json.Unmarshal(scanner.Bytes(), datum)
		if err != nil {
			return
		}
		line += 1

		id, ok := datum[idField]
		if !ok {
			err = fmt.Errorf("document at line %v does not have ID field "+
				"'%v' defined", line, idField)
			return
		}

		for _, field := range contentFields {
			content, ok := datum[field]
			if ok {
				fmt.Printf("[%v] -> [%v]", id, content)
				// err = b.AddDocument(id, content)
				if err != nil {
					return
				}
			}
		}

		fmt.Print(datum)
	}

	return
}

func (b *indexBuilderImpl) Dump(dst io.Writer) error {
	return b.trie.Dump(dst)
}
