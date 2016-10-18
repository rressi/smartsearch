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
	ScanJsonStream(reader io.Reader, idField string,
		contentFields []string) (numLines int, err error)
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
	contentFields []string) (numLines int, err error) {

	// Any further failure will reset our state machine:
	defer func() {
		if err == io.EOF {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("IndexBuilder.ScanJsonStream, line %d: %v",
				numLines, err)
		}
	}()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue // Ignores empty lines.
		}
		numLines += 1

		var datum map[string]interface{}
		err = json.Unmarshal(scanner.Bytes(), &datum)
		if err != nil {
			return
		}

		id_, ok := datum[idField]
		if !ok {
			err = fmt.Errorf("document at line %v does not have ID field "+
				"'%v' defined", numLines, idField)
			return
		}

		var id int
		id, err = strconv.Atoi(fmt.Sprint(id_))
		if err != nil {
			return
		}

		for _, field := range contentFields {
			content_, ok := datum[field]
			if ok {
				content := fmt.Sprint(content_)
				err = b.AddDocument(id, content)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (b *indexBuilderImpl) Dump(writer io.Writer) error {
	return b.trie.Dump(writer)
}
