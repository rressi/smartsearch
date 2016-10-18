package main

import (
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

	run(*inputFile, *outputFile, *jsonId, *jsonContents)
}

func run(
	inputFile string,
	outputFile string,
	jsonId string,
	jsonContents string) {

	// Handles feedback:
	fmt.Printf("input file: %v\n", inputFile)
	fmt.Printf("output file: %v\n", outputFile)
	fmt.Printf("json id: %v\n", jsonId)
	fmt.Printf("json contents: %v\n", jsonContents)
	var err error
	defer func() {
		if err == nil {
			fmt.Print("Done.\n")
		} else {
			fmt.Printf("Error: %v\n", err)
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

	// Serializes the index:
	err = builder.Dump(output)
	if err != nil {
		return
	}

	return
}
