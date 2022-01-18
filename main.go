package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const (
	stateNone = iota
	stateInCodeBlock
)

// Match: ![render-db6d08bb022ed12c2cc74d86d7a4707d.svg](/optional/path/to/render-db6d08bb022ed12c2cc74d86d7a4707d.svg)
var renderedImageRegexp = regexp.MustCompile(`!\[render-.{32}\..+\]\(.*render-.{32}\..+\)`)

type Args struct {
	Types      []string // Languages to render
	OutputDir  string   // Directory to output rendered files to
	LinkPrefix string   // Prefix to use when linking to rendered files
	Files      []string // Markdown files to process
}

func (a *Args) Parse() error {
	types := flag.String("types", "", "Languages to render (required, comma-separated, supported languages: [dot])")
	outputDir := flag.String("output-dir", "", "Directory to render code blocks to. If not specified, output will be rendered to the same directory as the input file.")
	linkPrefix := flag.String("link-prefix", "", "Prefix to use when linking to rendered files")
	flag.Parse()

	if *types == "" {
		return errors.New("--types is required")
	}
	if flag.NArg() == 0 {
		return errors.New("no files specified")
	}

	a.OutputDir = *outputDir
	a.LinkPrefix = *linkPrefix
	a.Types = strings.Split(*types, ",")
	a.Files = flag.Args()
	return nil
}

type CodeBlock struct {
	Type     string
	Fenced   string
	Contents string
}

func (c *CodeBlock) Reset() {
	c.Type = ""
	c.Fenced = ""
	c.Contents = ""
}

func main() {
	args := &Args{}
	if err := args.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, v := range args.Files {
		err := processFile(v, args.Types, args.OutputDir, args.LinkPrefix)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func processFile(filePath string, types []string, outputDir string, linkPrefix string) error {
	err := validateFileExists(filePath)
	if err != nil {
		return err
	}

	// If output directory is not specified, default to the directory of
	// the file being processed.
	if outputDir == "" {
		outputDir = path.Dir(filePath)
	}

	// Construct a lookup for O(1) access
	typeLookup := make(map[string]bool)
	for _, v := range types {
		typeLookup[v] = true
	}

	// Initialize state
	var lines []string
	currentCodeBlock := &CodeBlock{}
	state := stateNone
	hasChanges := false

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch state {
		case stateNone:
			if strings.HasPrefix(line, "```") {
				for k := range typeLookup {
					if strings.HasPrefix(line, fmt.Sprintf("```%s render", k)) {
						state = stateInCodeBlock
						currentCodeBlock.Fenced = line + "\n"
						currentCodeBlock.Type = k
						break
					}
				}
			}
			// Append lines if state didn't change
			if state == stateNone {
				lines = append(lines, line)
			}
		case stateInCodeBlock:
			// Check if we're exiting a rendered code block
			if strings.HasPrefix(line, "```") {
				state = stateNone
				currentCodeBlock.Fenced += line
				hasChanges = true

				// Render the code block
				outputFileName, err := renderCodeBlock(currentCodeBlock, outputDir)
				if err != nil {
					return err
				}
				fmt.Printf("file=%s rendered=%s\n", filePath, outputFileName)

				// Check if code block has been rendered before by looking at the previous two lines.
				isRenderedBefore := false
				for i := 1; i <= 2; i++ {
					prevLine := lines[len(lines)-i]
					if renderedImageRegexp.MatchString(prevLine) {
						lines[len(lines)-i] = buildMarkdownImage(outputFileName, linkPrefix)
						isRenderedBefore = true
						break
					}
				}

				// If not rendered before, add the image link above the code block
				if !isRenderedBefore {
					lines = append(lines, buildMarkdownImage(outputFileName, linkPrefix), "")
				}

				// Add the fenced code block back to the file
				lines = append(lines, strings.Split(currentCodeBlock.Fenced, "\n")...)

				currentCodeBlock.Reset()
				continue
			} else {
				// We're still within the code block, so continue adding collecting its lines.
				currentCodeBlock.Fenced += line + "\n"
				currentCodeBlock.Contents += line + "\n"
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write updated contents back to disk
	if hasChanges {
		writer, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer writer.Close()
		output := strings.Join(lines, "\n")
		writer.WriteString(output)
	}
	return nil
}

// validate that file exists and is not a directory
func validateFileExists(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", filePath)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("%s is a directory", filePath)
	}
	return nil
}

func runShellCommand(command string, args []string, stdin io.Reader) (stdoutOutput []byte, err error) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	err = cmd.Run()
	return stdout.Bytes(), err
}

func renderCodeBlock(codeBlock *CodeBlock, outputDir string) (fileName string, err error) {
	var content []byte
	var ext string

	switch codeBlock.Type {
	case "dot":
		content, err = runShellCommand("dot", []string{"-Tsvg"}, strings.NewReader(codeBlock.Contents))
		if err != nil {
			return "", err
		}
		ext = "svg"
	default:
		return "", fmt.Errorf("unsupported type: %s", codeBlock.Type)
	}

	fileHash := fmt.Sprintf("%x", md5.Sum([]byte(codeBlock.Contents)))
	fileName = "render-" + fileHash + "." + ext
	outputFilePath := path.Join(outputDir, fileName)
	f, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.Write(content)
	return fileName, nil
}

func buildMarkdownImage(outputFilename, linkPrefix string) string {
	return fmt.Sprintf("![%s](%s)", outputFilename, linkPrefix+outputFilename)
}
