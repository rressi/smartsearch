package smartsearch

import (
	"reflect"
	"testing"
)

func TestTokenizer_Base(t *testing.T) {

	tokenizer := NewTokenizer()

	var err error
	var tokens, expected_tokens []string
	var query string

	query = "YES!-This ìs ä fÄncy, is a string"
	expected_tokens = []string{
		"yes", "this", "is", "a", "fancy", "is", "a", "string"}

	tokens, err = tokenizer.Apply(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	}

	query = ""
	expected_tokens = nil

	tokens, err = tokenizer.Apply(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	}
}

func TestTokenizer_Search(t *testing.T) {

	tokenizer := NewTokenizer()

	var err error
	var tokens, expected_tokens []string
	var incomplete_token, expected_incomplete_token string
	var query string

	query = "YES!-This ìs ä fÄncy, is a string"
	expected_tokens = []string{"a", "fancy", "is", "this", "yes"}
	expected_incomplete_token = "string"

	tokens, incomplete_token, err = tokenizer.ForSearch(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	} else if incomplete_token != expected_incomplete_token {
		t.Errorf("Unexpected result: incomplete_token=%v", incomplete_token)
	}

	query = "YES!-This ìs ä fÄncy, is a string-"
	expected_tokens = []string{"a", "fancy", "is", "string", "this", "yes"}
	expected_incomplete_token = ""

	tokens, incomplete_token, err = tokenizer.ForSearch(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	} else if incomplete_token != expected_incomplete_token {
		t.Errorf("Unexpected result: incomplete_token=%v", incomplete_token)
	}

	query = ""
	expected_tokens = nil
	expected_incomplete_token = ""

	tokens, incomplete_token, err = tokenizer.ForSearch(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	} else if incomplete_token != expected_incomplete_token {
		t.Errorf("Unexpected result: incomplete_token=%v", incomplete_token)
	}

	query = "Th"
	expected_tokens = nil
	expected_incomplete_token = "th"

	tokens, incomplete_token, err = tokenizer.ForSearch(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	} else if incomplete_token != expected_incomplete_token {
		t.Errorf("Unexpected result: incomplete_token=%v", incomplete_token)
	}

	query = "TH "
	expected_tokens = []string{"th"}
	expected_incomplete_token = ""

	tokens, incomplete_token, err = tokenizer.ForSearch(query)
	if err != nil {
		t.Errorf("Tokenization has failed: %v", err)
	} else if !reflect.DeepEqual(tokens, expected_tokens) {
		t.Errorf("Unexpected result: tockens=%v", tokens)
	} else if incomplete_token != expected_incomplete_token {
		t.Errorf("Unexpected result: incomplete_token=%v", incomplete_token)
	}
}
