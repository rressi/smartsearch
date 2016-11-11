package smartsearch

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"testing"
)

var ATTRIBUTES = []string{
	"Business Account Number",
	"City",
	"DBA Name",
	"Location Id",
	"Mail Address",
	"Mail City",
	"Mail State",
	"Mail Zipcode",
	"NAICS Code",
	"NAICS Code Description",
	"Neighborhoods - Analysis Boundaries",
	"Ownership Name",
	"Source Zipcode",
	"State",
	"Street Address",
	"Supervisor District"}

var cachedDocs [][]byte

func loadInput() (docs [][]byte, err error) {

	if cachedDocs != nil {
		docs = cachedDocs
		return
	}
	println("Loading...")

	defer func() {
		if err != nil {
			err = fmt.Errorf("loadInput: %v", err)
		}
	}()

	// Opens the input file:
	var inputFile *os.File
	inputFile, err = os.Open(
		"../../../../data/business-locations-distilled.json.gz")
	if err != nil {
		return
	}
	defer inputFile.Close()

	// Decompresses it:
	var decoder io.Reader
	decoder, err = gzip.NewReader(inputFile)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(decoder)
	var docs_ [][]byte
	for scanner.Scan() {
		var buf bytes.Buffer
		buf.Write(scanner.Bytes())
		docs_ = append(docs_, buf.Bytes())
	}

	err = scanner.Err()
	if err != nil {
		return
	}

	docs = docs_
	cachedDocs = docs_
	fmt.Printf("%v documents loaded\n", len(docs))
	return
}

func benchmarkIndexBuilder(docs [][]byte) {

	// Handles errors:
	var err error
	defer func() {
		if err != nil {
			err = fmt.Errorf("benchmarkIndexBuilder(%v): %v", len(docs), err)
			panic(err)
		}
	}()

	// Creates an IndexBuilder
	builder := NewIndexBuilder()
	defer builder.Abort() // This protects us from leaking some go-routine

	// Reads it line by line:
	numInputBytes := 0
	for _, docBytes := range docs {

		numInputBytes += len(docBytes)
		builder.AddJsonDocument(docBytes, "uuid", ATTRIBUTES)
	}

	var buf bytes.Buffer
	builder.Dump(&buf)

	var zbuf bytes.Buffer
	encoder := gzip.NewWriter(&zbuf)
	encoder.Write(buf.Bytes())
	encoder.Close()

	fmt.Printf("\n%v documents: %v -> %v,%v: %0.2f%%,%0.2f%%\n",
		len(docs), numInputBytes, buf.Len(), zbuf.Len(),
		100.0*float64(buf.Len())/float64(numInputBytes),
		100.0*float64(zbuf.Len())/float64(numInputBytes))
}

func BenchmarkIndexBuilder10000(b *testing.B) {
	docs, err := loadInput()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkIndexBuilder(docs[:10000])
	}
}

func BenchmarkIndexBuilder50000(b *testing.B) {
	docs, err := loadInput()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkIndexBuilder(docs[:50000])
	}
}

func BenchmarkIndexBuilder100000(b *testing.B) {
	docs, err := loadInput()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkIndexBuilder(docs[:100000])
	}
}

func BenchmarkIndexBuilder200000(b *testing.B) {
	docs, err := loadInput()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkIndexBuilder(docs[:200000])
	}
}
