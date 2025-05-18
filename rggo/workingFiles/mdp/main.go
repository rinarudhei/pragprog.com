package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
<h1>filename: {{ .Filename}}</h1>
{{ .Body }}
</body>
</html>
`
)

type content struct {
	Title    string
	Body     template.HTML
	Filename string
}

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip preview the html content")
	tFname := flag.String("t", "", "Alternate template name")
	interactive := flag.Bool("i", false, "Use interactive mode to add user input")
	flag.Parse()

	if *filename == "" && !*interactive {
		flag.Usage()
		os.Exit(1)
	}
	if *filename == "" && *interactive {
		fmt.Println("enter markdown file name:")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("processing markdown file: %s", text)
		*filename = strings.TrimSpace(text)
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

	htmlData, err := parseContent(b, tFName, filename)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	tmpFile, err := os.CreateTemp("./", "mdp*.html")
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
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
func parseContent(mdBytesData []byte, tFname string, filename string) ([]byte, error) {
	out := blackfriday.Run(mdBytesData)
	body := bluemonday.UGCPolicy().SanitizeBytes(out)
	tmp, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parseContent: %w", err)
	}
	if tFname != "" {
		tmp, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, fmt.Errorf("parseContent: parse files: %w", err)
		}
	} else if os.Getenv("HTML_TEMPLATE_ENV") != "" {
		tFname = os.Getenv("HTML_TEMPLATE_ENV")
		tmp, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, fmt.Errorf("parseContent: parse files from env: %w", err)
		}
	}
	content := content{
		Title:    "Markdown Preview Tool",
		Body:     template.HTML(body),
		Filename: filename,
	}
	var buf bytes.Buffer
	if err := tmp.Execute(&buf, content); err != nil {
		return nil, fmt.Errorf("parseContent: execute: %w", err)
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

	time.Sleep(2 * time.Second)

	return nil
}
