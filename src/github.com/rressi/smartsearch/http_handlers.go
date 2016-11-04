package smartsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Creates a http.Handler to serve the passed collection of JSON documents.
//
// Passed collection is a map with the document uuid as a key (integer) and
// the raw JSON content to return as value.
//
// The web API exposed by this handler accept the following arguments:
// - ids: a space separated list of documents' uuids to select the documents
//   to be returned. They are returned in the very same order ar respective
//   uuids in this parameter.
// - l: a positive integer to limit the number of returned document.
//
// Notes:
// - If argument "ids" is not passed all the documents are returned sorted by
//   uuids.
// - If just one document uuid passed with argument "ids" is not valid this web
//   request fails.
//
// This handler returns as a content one text file with one document per line
// encoded in JSON format (the same raw bytes of the passed collection of
// documents passed originally).
func ServeDocuments(docs JsonDocuments) http.Handler {

	// Obtains all ids:
	allIds := make([]int, 0, len(docs))
	for k := range docs {
		allIds = append(allIds, k)
	}
	sort.Ints(allIds)

	// Our web handler:
	docsHandler := func(w http.ResponseWriter, r *http.Request) {

		var httpError = http.StatusInternalServerError
		var err error
		defer func() {
			if err != nil {
				fmt.Printf("Error: ServeDocuments: %v\n", err)
				err = fmt.Errorf("ServeDocuments: %v", err)
				if httpError != 0 {
					w.WriteHeader(httpError)
				}
			}
		}()

		var values map[string][]string
		values, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			httpError = http.StatusBadRequest
			return
		}

		var limit int
		limit, err = parseNumericalArgument("l", values)
		if err != nil {
			httpError = http.StatusBadRequest
			return
		}

		var selectedIds []int
		idsValuesMulti, idsOk := values["ids"]
		if idsOk {
			// Parses all passed ids and checks their validity:
		parseLoop:
			for _, idsValues := range idsValuesMulti {
				for _, idRaw := range strings.Split(idsValues, " ") {
					if limit >= 0 && len(selectedIds) >= limit {
						break parseLoop
					}

					var id int
					id, err = strconv.Atoi(idRaw)
					if err != nil {
						err = fmt.Errorf("non numeric id: '%v'", idRaw)
						httpError = http.StatusBadRequest
						return
					}

					_, docOk := docs[id]
					if !docOk {
						err = fmt.Errorf("invalid document id: %v", id)
						httpError = http.StatusNotFound
						return
					}

					selectedIds = append(selectedIds, id)
				}
			}
			if limit != 0 && len(selectedIds) == 0 {
				err = errors.New("No document ids have been passed")
				httpError = http.StatusBadRequest
				return
			}
		} else {
			selectedIds = allIds
			if limit >= 0 && len(selectedIds) > limit {
				selectedIds = selectedIds[:limit]
			}
		}

		w.WriteHeader(http.StatusOK)
		httpError = 0
		// NOTE: it is no more possible to return an error to the client.

		var count int
		// Writes back all the documents...
		for _, id := range selectedIds {
			if limit >= 0 && count > limit {
				break
			}

			rawDocument, idsOk := docs[id]
			if !idsOk {
				err = fmt.Errorf("invalid documente id: %v", id)
				return
			}

			_, err = w.Write(rawDocument)
			if err != nil {
				return
			}

			_, err = w.Write([]byte{'\n'})
			if err != nil {
				return
			}
		}
	}

	return http.HandlerFunc(docsHandler)
}

// -----------------------------------------------------------------------------

// Creates an http.Handler to search using the given index.
func ServeSearch(index Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var httpError = http.StatusInternalServerError
		var err error
		defer func() {
			if err != nil {
				err = fmt.Errorf("ServeSearch: %v", err)
				if httpError != 0 {
					w.WriteHeader(httpError)
				}
			}
		}()

		var values map[string][]string
		values, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			httpError = http.StatusBadRequest
			return
		}

		var query string
		queryValues, queryOk := values["q"]
		if !queryOk {
			// pass
		} else if len(queryValues) != 1 {
			httpError = http.StatusBadRequest
			err = errors.New("Parameter 'q' passed more than once")
			return
		} else {
			query = queryValues[0]
		}

		var limit int
		limit, err = parseNumericalArgument("l", values)
		if err != nil {
			httpError = http.StatusBadRequest
			return
		}

		var postings []int
		postings, err = index.Search(query, limit)
		if err != nil {
			httpError = http.StatusNotFound
			return
		} else if postings == nil {
			postings = make([]int, 0)
		}

		var buf []byte
		buf, err = json.Marshal(postings)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf)
		if err != nil {
			return
		}

		httpError = 0 // Done!
	}
}

// -----------------------------------------------------------------------------

// Just servers passed bytes via http.
func ServeRawBytes(raw []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var httpError = http.StatusInternalServerError
		var err error
		defer func() {
			if err != nil {
				err = fmt.Errorf("ServeRawBytes: %v", err)
				if httpError != 0 {
					w.WriteHeader(httpError)
				}
			}
		}()

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(raw)
		if err != nil {
			return
		}

		httpError = 0 // Done!
	}
}

// Parses an argument of type integer value from an HTTP request.
func parseNumericalArgument(name string, values map[string][]string) (
	limit int, err error) {

	var limit_ int
	limitValues, limitOk := values[name]
	if !limitOk {
		limit_ = -1
	} else if len(limitValues) != 1 {
		err = errors.New("Parameter 'limit' passed more than once")
		return
	} else {
		limit_, err = strconv.Atoi(limitValues[0])
		if err != nil {
			err = fmt.Errorf("invalid value for parameter 'limit': %v",
				limitValues[0])
			return
		}
	}

	// Success!
	limit = limit_
	return
}
