package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<title>{{ .Title }}</title>
</head>
<body>
{{ .Body }}
</body>
</html>
`
)

type content struct {
	Title string
	Body  template.HTML
}

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip preview the html content")
    tFname := flag.String("t", "", "Alternate template name")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *tFname, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Generate sanitized html file from markdown file data
func run(filename string, tFName string, out io.Writer, skipPreview bool) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	htmlData, err := parseContent(b, tFName)
    if err != nil {
        return fmt.Errorf("run: %w", err)
    }
	tmpFile, err := os.CreateTemp("./", "mdp*.html")
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	tmpFile.Close()
	outName := tmpFile.Name()
	fmt.Fprintln(out, outName)

	if err := saveHtml(outName, htmlData); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	if skipPreview {
		return nil
	}
	defer os.Remove(outName)

	return preview(outName)
}

// Parse markdown file content into sanitized html content
func parseContent(mdBytesData []byte, tFname string) ([]byte, error) {
	out := blackfriday.Run(mdBytesData)
	body := bluemonday.UGCPolicy().SanitizeBytes(out)
	tmp, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parseContent: %w", err)
	}
    if tFname != "" {
        tmp, err = template.New("mdp").ParseFiles(tFname)
    }
	content := content{
		Title: "Markdown Preview Tool",
		Body:  template.HTML(body),
	}
    var buf bytes.Buffer
    if err := tmp.Execute(&buf, content); err != nil {
        return nil, fmt.Errorf("parseContent: %w", err)
    }

   return buf.Bytes(), nil
}

// Generate html file output
func saveHtml(outName string, htmlData []byte) error {
	if err := os.WriteFile(outName, htmlData, 0644); err != nil {
		return fmt.Errorf("saveHtml: %w", err)
	}

	return nil
}

// Preview html content
func preview(fname string) error {
	cName := ""
	cParams := []string{}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return errors.New("preview: OS not supported")
	}

	cParams = append(cParams, fname)
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return fmt.Errorf("preview: %w", err)
	}

	if err := exec.Command(cPath, cParams...).Run(); err != nil {
		return fmt.Errorf("preview: %w", err)
	}

	time.Sleep(5 * time.Second)

	return nil
}
