package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/rressi/smartsearch"
	"io"
	"os"
	"strings"
)

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

func runMakeIndex(
	inputFile string,
	outputFile string,
	jsonId string,
	jsonContents string) {

	// Handles feedback:
	fmt.Fprintf(os.Stderr, "input file: %v\n", inputFile)
	fmt.Fprintf(os.Stderr, "output file: %v\n", outputFile)
	fmt.Fprintf(os.Stderr, "json id: %v\n", jsonId)
	fmt.Fprintf(os.Stderr, "json contents: %v\n", jsonContents)
	var err error
	defer func() {
		if err == nil {
			fmt.Fprint(os.Stderr, "Done.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}()

	// Selects the input:
	var input io.Reader
	if inputFile == "-" {
		input = os.Stdin
	} else {
		input, err = os.Open(inputFile)
		if err != nil {
			return
		}
	}

	// We prefer to have buffered I/0:
	input = bufio.NewReader(input)

	// Indexes all the documents:
	builder := smartsearch.NewIndexBuilder()
	jsonContentsSplit := strings.Split(jsonContents, ",")
	err = builder.ScanJsonStream(input, jsonId, jsonContentsSplit)
	if err != nil {
		return
	}

	// Selects the output:
	var output io.Writer
	if outputFile == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputFile)
		if err != nil {
			return
		}
	}

	// We prefer to have buffered I/0:
	output = bufio.NewWriter(output)

	// Serializes the index:
	err = builder.Dump(output)
	if err != nil {
		return
	}

	return
}
