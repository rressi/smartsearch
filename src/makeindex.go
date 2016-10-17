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
	jsonId := flags.String("id", "id", "Json attribute for doc ids")
	jsonContent := flags.String("content", "content",
		"Json attributes to be indexed, comma separated")
	sourceFiles := flags.String("source", "-", "Soure file(s) to be read, use"+
		" comma as a separator")
	outputFile := flags.String("o", "-", "Output file.")
	err = flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		flag.Usage()
		return
	}

	fmt.Printf("source file: %v\n", *sourceFiles)
	fmt.Printf("json id: %v\n", *jsonId)
	fmt.Printf("json content: %v\n", *jsonContent)
	fmt.Printf("output file: %v\n", *outputFile)

	run(*sourceFiles, *outputFile, *jsonId, *jsonContent)
}

func run(
	sourceFiles string,
	outputFIle string,
	jsonId string,
	jsonContent string) {

	// Handles errors:

	var err error
	defer func() {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Reads all the source files:

	builder := smartsearch.NewIndexBuilder()
	if sourceFiles == "-" {
		err = builder.ScanJsonStream(os.Stdin, jsonId,
			strings.Split(jsonContent, ","))
		if err != nil {
			return
		}
	} else {
		for _, sourceFile := range strings.Split(sourceFiles, ",") {

			var source io.Reader
			source, err = os.Open(sourceFile)
			if err != nil {
				return
			}

			err = builder.ScanJsonStream(source, jsonId,
				strings.Split(jsonContent, ","))
			if err != nil {
				return
			}
		}
	}

	// Produces the output:

	var output io.Writer
	if outputFIle == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputFIle)
		if err != nil {
			return
		}
	}

	err = builder.Dump(output)
	if err != nil {
		return
	}

	return
}
