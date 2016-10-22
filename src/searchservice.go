package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/rressi/smartsearch"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"strconv"
)

var index smartsearch.Index
var documents smartsearch.JsonDocuments

func main() {

	var err error

	// Parses command line parameters:
	flags := flag.NewFlagSet("searchservice", flag.ExitOnError)
	documentsFile := flags.String("d", "", "File containing all the documents")
	indexFile := flags.String("i", "-", "Raw index as input file")
	httpHostName := flags.String("n", "", "Optional HTTP host name.")
	httpPort := flags.Uint("p", 5000, "TCP port to be used by the HTTP server.")
	jsonId := flags.String("id", "id", "Json attribute for document ids")
	jsonContents := flags.String("content", "content",
		"Json attributes to be indexed, comma separated")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		flag.Usage()
		return
	}

	// Handles feedback to the user:
	fmt.Fprint(os.Stderr, "[searchservice]\n")
	if *documentsFile != "" {
		fmt.Fprintf(os.Stderr, "Documents file: %v\n", *indexFile)
	} else if indexFile != nil {
		fmt.Fprintf(os.Stderr, "input file: %v\n", *indexFile)
	}
	fmt.Fprintf(os.Stderr, "http host name: %v\n", *httpHostName)
	fmt.Fprintf(os.Stderr, "http port: %v\n", *httpPort)
	defer func() {
		if err == nil {
			fmt.Fprint(os.Stderr, "Done.\n\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		}
	}()

	if *documentsFile != "" {
		documents, index, err = LoadDocuments(*documentsFile, *jsonId,
			*jsonContents)
	} else {
		documents = nil
		index, err = IndexDocuments(*indexFile)
	}
	if err != nil {
		return
	}

	// Executes our service:
	fmt.Fprint(os.Stderr, "listening...\n")
	err = RunSearchService(*httpHostName, *httpPort)
	if err != nil {
		return
	}
}

func LoadDocuments(documentFile string,
	jsonId string,
	jsonContents string) (
	documents smartsearch.JsonDocuments, index smartsearch.Index, err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("LoadDocuments: %v", err)
		}
	}()

	// Selects the input:
	var input io.Reader
	if documentFile == "-" {
		input = os.Stdin
	} else {
		var fileInput io.ReadCloser
		fileInput, err = os.Open(documentFile)
		if err != nil {
			return
		}
		defer fileInput.Close()
		input = fileInput
	}

	// We prefer to have buffered I/0:
	bufInput := bufio.NewReader(input)

	// Loads and indexes all the documents:
	builder := smartsearch.NewIndexBuilder()
	jsonContentsSplit := strings.Split(jsonContents, ",")
	documents, err = builder.LoadAndIndexJsonStream(bufInput, jsonId,
		jsonContentsSplit)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "documents loaded: %v\n", len(documents))

	indexBytes := new(bytes.Buffer)
	builder.Dump(indexBytes)
	index, err = smartsearch.NewIndex(indexBytes)
	return
}

func IndexDocuments(inputFile string) (index smartsearch.Index, err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("SetupSearchService: %v", err)
		}
	}()

	// Selects the input stream:
	var input io.ReadCloser
	if inputFile == "-" {
		input = os.Stdin
	} else {
		input, err = os.Open(inputFile)
		defer input.Close()
		if err != nil {
			return
		}
	}

	// Loads the index from the input stream:
	index, err = smartsearch.NewIndex(input)
	return
}

func RunSearchService(httpHostName string, httpPort uint) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("RunSearchService: %v", err)
		}
	}()

	if index == nil {
		err = errors.New("Index not loaded")
		return
	}

	// Creates the web server and listen for incoming requests:
	if documents != nil {
		http.HandleFunc("/getDocument", DocumentsHandler)
	}
	http.HandleFunc("/search", SearchHandler)
	address := fmt.Sprintf("%v:%v", httpHostName, httpPort)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		return
	}

	return
}

func DocumentsHandler(w http.ResponseWriter, r *http.Request) {

	var httpError = http.StatusInternalServerError
	var err error
	defer func() {
		if err != nil {
			err = fmt.Errorf("DocumentsHandler: %v", err)
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

	ids, ok := values["ids"]
	if !ok || len(ids) == 0 {
		httpError = http.StatusBadRequest
		err = errors.New("Missing parameter 'ids'")
		return
	}

	w.WriteHeader(http.StatusOK)
	for _, idRaw := range strings.Split(ids[0], " ") {
		var id int
		id, err = strconv.Atoi(idRaw)
		if err != nil {
			httpError = http.StatusBadRequest
			return
		}

		var rawDocument []byte
		rawDocument, ok = documents[id]
		if !ok {
			httpError = http.StatusNotFound
			err = fmt.Errorf("invalid id %v", id)
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

	httpError = 0 // Done!
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {

	var httpError = http.StatusInternalServerError
	var err error
	defer func() {
		if err != nil {
			err = fmt.Errorf("SearchHandler: %v", err)
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

	query, ok := values["q"]
	if !ok || len(query) == 0 {
		httpError = http.StatusBadRequest
		err = errors.New("Missing parameter 'q'")
		return
	}

	var postings []int
	postings, err = index.Search(query[0])
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
