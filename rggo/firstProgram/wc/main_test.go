package main

import (
	"bytes"
	"testing"
)

func TestCountWords(t *testing.T) {
	b := bytes.NewBufferString("word1 word2 word3 word4\n")
	exp := 4

	res := count(b, false, false)

	if res != exp {
		t.Fatalf("actual: %d, expected, %d", res, exp)
	}
}

func TestCountLines(t *testing.T) {
	b := bytes.NewBufferString("word1 word2\n word3 \nword4\n")
	exp := 3

	res := count(b, true, false)

	if res != exp {
		t.Fatalf("actual: %d, expected, %d", res, exp)
	}
}

func TestCountBytes(t *testing.T) {
	b := bytes.NewBufferString("word1 word2\n word3 \nword4\n")
	exp := b.Len()

	res := count(b, false, true)

	if res != exp {
		t.Fatalf("actual: %d, expected, %d", res, exp)
	}
}
