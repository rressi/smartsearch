package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/rressi/smartsearch"
	"io"
	"net/http"
	"os"
	"strings"
)

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
	staticAppFolder := flags.String("app", "", "optionally serves a static web"+
		" app from this passed folder")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		flag.Usage()
		return
	}

	// Handles feedback to the user:
	fmt.Fprint(os.Stderr, "[searchservice]\n")
	if *documentsFile != "" {
		fmt.Fprintf(os.Stderr, "Documents file:     %v\n", *documentsFile)
		fmt.Fprintf(os.Stderr, "Id attribute:       %v\n", *jsonId)
		fmt.Fprintf(os.Stderr, "Content attributes: %v\n", *jsonContents)
	}
	if *indexFile != "" {
		fmt.Fprintf(os.Stderr, "input file:         %v\n", *indexFile)
	}
	if *staticAppFolder != "" {
		fmt.Fprintf(os.Stderr, "app folder:         %v\n", *staticAppFolder)
	}
	fmt.Fprintf(os.Stderr, "http host name:     %v\n", *httpHostName)
	fmt.Fprintf(os.Stderr, "http port:          %v\n", *httpPort)
	defer func() {
		if err == nil {
			fmt.Fprint(os.Stderr, "Done.\n\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		}
	}()

	var ctx AppContext
	if *documentsFile != "" {
		ctx, err = LoadDocuments(*documentsFile, *jsonId, *jsonContents)
	} else {
		ctx, err = LoadIndex(*indexFile)
	}
	if err != nil {
		return
	}

	if *staticAppFolder != "" {
		ctx.staticAppFolder = *staticAppFolder
	}

	// Executes our service:
	fmt.Fprint(os.Stderr, "listening...\n")
	err = RunSearchService(ctx, *httpHostName, *httpPort)
	if err != nil {
		return
	}
}

// -----------------------------------------------------------------------------

type AppContext struct {
	docs            smartsearch.JsonDocuments
	rawIndex        []byte
	index           smartsearch.Index
	staticAppFolder string
}

func LoadDocuments(documentFile string, jsonId string, jsonContents string) (
	ctx AppContext, err error) {

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
	ctx.docs, err = builder.LoadAndIndexJsonStream(bufInput, jsonId,
		jsonContentsSplit)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "documents loaded: %v\n", len(ctx.docs))

	indexBytes := new(bytes.Buffer)
	builder.Dump(indexBytes)
	ctx.index, ctx.rawIndex, err = smartsearch.NewIndex(indexBytes)
	return
}

func LoadIndex(inputFile string) (ctx AppContext, err error) {

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
	ctx.index, ctx.rawIndex, err = smartsearch.NewIndex(input)
	return
}

func RunSearchService(ctx AppContext, httpHostName string, httpPort uint) (
	err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("RunSearchService: %v", err)
		}
	}()

	if ctx.index == nil {
		err = errors.New("Index not loaded")
		return
	}

	// Creates the web server and listen for incoming requests:
	http.HandleFunc("/search", smartsearch.ServeSearch(ctx.index))
	http.HandleFunc("/rawIndex", smartsearch.ServeRawBytes(ctx.rawIndex))
	if ctx.docs != nil {
		http.HandleFunc("/docs", smartsearch.ServeDocuments(ctx.docs))
	}
	if ctx.staticAppFolder != "" {
		http.Handle("/app/", http.StripPrefix("/app/",
			http.FileServer(http.Dir(ctx.staticAppFolder))))

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/app", http.StatusSeeOther)
		})
	}

	address := fmt.Sprintf("%v:%v", httpHostName, httpPort)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		return
	}

	return
}

// -----------------------------------------------------------------------------
