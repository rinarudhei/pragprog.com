package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func filterOut(path, ext string, minSize int64, info fs.FileInfo) bool {
	if info.IsDir() || info.Size() < minSize {
		return true
	}

	if ext != "" && ext != filepath.Ext(path) {
		return true
	}

	return false
}

func listFile(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}

func defFile(path string, delLogger *log.Logger) error {
	if err := os.Remove(path); err != nil {
		return err
	}

	delLogger.Println(path)

	return nil
}
