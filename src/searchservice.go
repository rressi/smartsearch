package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/rressi/smartsearch"
	"io"
	"net/http"
	"net/url"
	"os"
)

var index smartsearch.Index

func main() {

	var err error

	// Parses command line parameters:
	flags := flag.NewFlagSet("searchservice", flag.ExitOnError)
	inputFile := flags.String("i", "-", "Raw index as input file")
	httpHostName := flags.String("n", "", "Optional HTTP host name.")
	httpPort := flags.Uint("p", 5000, "TCP port to be used by the HTTP server.")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		flag.Usage()
		return
	}

	// Handles feedback to the user:
	fmt.Fprint(os.Stderr, "[searchservice]\n")
	fmt.Fprintf(os.Stderr, "input file: %v\n", *inputFile)
	fmt.Fprintf(os.Stderr, "http host name: %v\n", *httpHostName)
	fmt.Fprintf(os.Stderr, "http port: %v\n", *httpPort)
	defer func() {
		if err == nil {
			fmt.Fprint(os.Stderr, "Done.\n\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		}
	}()

	// Loads the index to memory:
	index, err = SetupSearchService(*inputFile)
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

func SetupSearchService(inputFile string) (index smartsearch.Index, err error) {

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

	// Creates the web server and listen for incoming requests:
	http.HandleFunc("/search", SearchHandler)
	address := fmt.Sprintf("%v:%v", httpHostName, httpPort)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		return
	}

	return
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
		err = errors.New("Missing query parameter 'q'")
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
