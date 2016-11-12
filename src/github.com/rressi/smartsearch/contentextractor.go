package smartsearch

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// A function to preprocess content in the slave threads
type ContentExtractor func(raw []byte) (id int, content string, err error)

func MakeJsonExtractor(idField string,
	contentFields []string) ContentExtractor {
	return func(jsonDocument []byte) (id int, content string, err error) {

		var datum map[string]interface{}
		err = json.Unmarshal(jsonDocument, &datum)
		if err != nil {
			return
		}

		var value interface{}
		value, ok := datum[idField]
		if !ok {
			err = fmt.Errorf("document does not have ID field '%v' defined",
				idField)
			return
		}

		// Parses the document id:
		var parsedId int
		switch docId_ := value.(type) {
		case int:
			parsedId = docId_
		case float64:
			parsedId = int(docId_)
		case string:
			parsedId, err = strconv.Atoi(docId_)
		}
		if err != nil {
			return
		}

		// Takes all the fields to be indexed:
		var parsedContent []string
		for _, field := range contentFields {
			value_, ok := datum[field]
			if ok {
				switch value := value_.(type) {
				case string:
					parsedContent = append(parsedContent, value)
				case int:
					parsedContent = append(parsedContent, fmt.Sprint(value))
				}
			}
		}

		id = parsedId
		content = strings.Join(parsedContent, " ")
		return
	}
}
