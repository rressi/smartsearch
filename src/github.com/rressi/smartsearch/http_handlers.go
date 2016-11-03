package smartsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Creates an http.Handler to serve the passed collection of JSON documents.
//
// Passed collection is a map with the document uuid as a key (integer) and
// the raw JSON content to return as value.
//
// The web API exposed by the created web server expects to have the following
// parameters:
// ids: it is mandatory and a space separated list of document uuid.
//
// The handle returns as a content one text file with one document per line
// encoded in JSON format (the same raw bytes of the passed collection of
// documents passed originally).
//
// Returned documents are the same requested with web parameter ids, in the very
// same order. Repeating many times the same ids just means to have the same
// document returned more than once.
//
// If one document uuid is not valid the web request fails.
func ServeDocuments(docs JsonDocuments) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

		idsValues, idsOk := values["ids"]
		if !idsOk || len(idsValues) == 0 {
			httpError = http.StatusBadRequest
			err = errors.New("Missing parameter 'ids'")
			return
		}

		w.WriteHeader(http.StatusOK)
		httpError = 0 // Done!
		for _, ids := range idsValues {
			for _, idRaw := range strings.Split(ids, " ") {
				var id int
				id, err = strconv.Atoi(idRaw)
				if err != nil {
					err = fmt.Errorf("non numeric id: '%v'", idRaw)
					httpError = http.StatusBadRequest
					return
				}

				var rawDocument []byte
				rawDocument, idsOk = docs[id]
				if !idsOk {
					httpError = http.StatusNotFound
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
	})
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
		limitValues, limitOk := values["l"]
		if !limitOk {
			limit = -1
		} else if len(limitValues) != 1 {
			httpError = http.StatusBadRequest
			err = errors.New("Parameter 'limit' passed more than once")
			return
		} else {
			limit, err = strconv.Atoi(limitValues[0])
			if err != nil {
				err = fmt.Errorf("invalid value for parameter 'limit': %v",
					limitValues[0])
				return
			}
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
