package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

const (
	inputFile  = "./testdata/test1.md"
	goldenFile = "./testdata/test1.md.html"
)

func TestParseContent(t *testing.T) {
	b, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}
	gold, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := parseContent(b, "")
    if err != nil {
        t.Fatal(err)
    }

	if !bytes.Equal(parsed, gold) {
		t.Logf("actual:\n %s\n", parsed)
		t.Logf("expected:\n %s\n", gold)
		t.Fatal("Result content does not match with expected content")
	}
}

func TestRun(t *testing.T) {
	var mockStdOut bytes.Buffer
	if err := run(inputFile, "", &mockStdOut, true); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(&mockStdOut)
	if err != nil {
		t.Fatal(err)
	}
	outputFile := strings.TrimSpace(string(output))
	result, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	gold, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, gold) {
		t.Logf("actual:\n %s\n", result)
		t.Logf("expected:\n %s\n", gold)
		t.Fatal("Result content does not match with expected content")
	}

	os.Remove(outputFile)
}
