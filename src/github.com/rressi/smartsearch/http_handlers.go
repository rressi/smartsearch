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

func ServeDocuments(docs JsonDocuments) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var httpError = http.StatusInternalServerError
		var err error
		defer func() {
			if err != nil {
				err = fmt.Errorf("AppDoc: %v", err)
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

		httpError = 0 // Done!
	}
}

// -----------------------------------------------------------------------------

func ServeSearch(index Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var httpError = http.StatusInternalServerError
		var err error
		defer func() {
			if err != nil {
				err = fmt.Errorf("AppSearch: %v", err)
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
				err = fmt.Errorf("AppRawBytes: %v", err)
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
