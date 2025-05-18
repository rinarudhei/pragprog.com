package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	countLines := flag.Bool("l", false, "count lines")
	countBytes := flag.Bool("b", false, "count bytes")
	file := flag.Bool("f", false, "scan text file")

	flag.Parse()
	var r io.Reader
	if *file {
		filenames := flag.Args()
		joinedByte := []byte{}
		for _, f := range filenames {
			b, err := os.ReadFile(f)
			if err != nil {
				fmt.Printf("Error reading a file: %s", f)
				os.Exit(1)
			}
			joinedByte = append(joinedByte, b...)
		}

		r = bytes.NewReader(joinedByte)
	} else {
		r = os.Stdin
	}

	fmt.Println(count(r, *countLines, *countBytes))
}

func count(r io.Reader, countLines bool, countBytes bool) int {
	scanner := bufio.NewScanner(r)
	if !countLines && !countBytes {
		scanner.Split(bufio.ScanWords)
	} else if countBytes {
		scanner.Split(bufio.ScanBytes)
	}

	wc := 0
	for scanner.Scan() {
		wc++
	}

	return wc
}
