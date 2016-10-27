package main

// Please read 'doc/makeindex.md' to know more about this tool.

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/rressi/smartsearch"
	"io"
	"os"
	"strings"
)

// Executable's main function.
func main() {
	var err error

	flags := flag.NewFlagSet("makeindex", flag.ExitOnError)
	inputFile := flags.String("i", "-", "Input file")
	outputFile := flags.String("o", "-", "Output file")
	jsonId := flags.String("id", "id", "Json attribute for document ids")
	jsonContents := flags.String("content", "content",
		"Json attributes to be indexed, comma separated")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		flag.Usage()
		return
	}

	runMakeIndex(*inputFile, *outputFile, *jsonId, *jsonContents)
}

// Takes as input a file with a stream of JSON documents and generates an index
// that it saves on an output file.
//
// Parameters:
// - inputFile:    A text file containing a stream of JSON documents, one per
//   line.
// - outputFile:   Target file where a binary index to be generated and dumped.
// - jsonId:       Attribute from the JSON document containing an id that is
//                 unique and mandatory for each document.
// - jsonContents: A list of top level attributes in each document whose
//                 values need to be indexed. It is ok if a document miss
//                 some or all of this attributes.
func runMakeIndex(
	inputFile string,
	outputFile string,
	jsonId string,
	jsonContents string) {

	// Handles feedback:
	fmt.Fprint(os.Stderr, "[makeindex]\n")
	fmt.Fprintf(os.Stderr, "input file: %v\n", inputFile)
	fmt.Fprintf(os.Stderr, "output file: %v\n", outputFile)
	fmt.Fprintf(os.Stderr, "json id: %v\n", jsonId)
	fmt.Fprintf(os.Stderr, "json contents: %v\n", jsonContents)
	var err error
	defer func() {
		if err == nil {
			fmt.Fprint(os.Stderr, "Done.\n\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		}
	}()

	// Selects the input:
	var input io.Reader
	if inputFile == "-" {
		input = os.Stdin
	} else {
		var fileInput io.ReadCloser
		fileInput, err = os.Open(inputFile)
		if err != nil {
			return
		}
		defer fileInput.Close()
		input = fileInput
	}

	// We prefer to have buffered I/0:
	bufInput := bufio.NewReader(input)

	// Indexes all the documents:
	var numLines int
	builder := smartsearch.NewIndexBuilder()
	jsonContentsSplit := strings.Split(jsonContents, ",")
	numLines, err = builder.IndexJsonStream(bufInput, jsonId, jsonContentsSplit)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "lines indexed: %v\n", numLines)

	// Selects the output:
	var output io.Writer
	if outputFile == "-" {
		output = os.Stdout
	} else {
		var outputF io.WriteCloser
		outputF, err = os.Create(outputFile)
		if err != nil {
			return
		}
		defer outputF.Close()
		output = outputF
	}

	// We prefer to have buffered I/0:
	bufOutput := bufio.NewWriter(output)
	defer bufOutput.Flush()

	// Serializes the index:
	err = builder.Dump(bufOutput)
	if err != nil {
		return
	}

	return
}
