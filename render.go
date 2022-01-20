package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Match: ![render-db6d08bb022ed12c2cc74d86d7a4707d.svg](/optional/path/to/render-db6d08bb022ed12c2cc74d86d7a4707d.svg)
// Capture group on the hash.
var renderedImageRegexp = regexp.MustCompile(`!\[render-.{32}\..+\]\(.*render-(.{32})\..+\)`)

// Chunk represents a segment of a file
type Chunk struct {
	Lines          []string // Lines the chunk contains
	StartLineIndex int      // Index is relative to the input file
	EndLineIndex   int      // Index is relative to the input file

	IsRenderable           bool
	Language               string
	ImageRelativeLineIndex int      // Where the image is located in the chunk. Index is relative to the chunk's lines.
	RenderedHash           string   // If image has been rendered before, contains the hash of the code block previously used to render the image
	CodeBlockContent       []string // The contents of the code block
}

func (r *Chunk) ShouldRender() bool {
	if !r.IsRenderable {
		return false
	}
	if r.HashContent() != r.RenderedHash {
		return true
	}
	return false
}

func (r *Chunk) HashContent() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(r.CodeBlockContent, "\n"))))
}

func (r *Chunk) Render(outputDir string, linkPrefix string) (fileName string, err error) {
	var content []byte
	var ext string

	switch r.Language {
	case "dot":
		content, err = runShellCommand("dot", []string{"-Tsvg"}, strings.NewReader(strings.Join(r.CodeBlockContent, "\n")))
		if err != nil {
			return "", err
		}
		ext = "svg"
	default:
		return "", fmt.Errorf("unsupported type: %s", r.Language)
	}

	fileName = "render-" + r.HashContent() + "." + ext
	outputFilePath := path.Join(outputDir, fileName)
	f, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.Write(content)

	// Update the chunk's lines
	image := buildMarkdownImage(fileName, linkPrefix)
	if r.RenderedHash != "" {
		r.Lines[r.ImageRelativeLineIndex] = image
	} else {
		r.Lines = append([]string{image, ""}, r.Lines...)
	}

	return fileName, nil
}

func NewRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render code blocks in markdown files",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no files specified as input")
			}
			return nil
		},
		RunE: renderCmd,
	}
	cmd.Flags().StringVar(&config.Render.OutputDir, "output-dir", "", "Directory to render code blocks to. If not specified, output will be rendered to the same directory as the input file.")
	cmd.Flags().StringVar(&config.Render.Languages, "languages", "", "Languages to render (comma-separated, supported languages: [dot])")
	cmd.MarkFlagRequired("languages")
	cmd.Flags().StringVar(&config.Render.LinkPrefix, "link-prefix", "", "Prefix to use when linking to rendered files")
	return cmd
}

func renderCmd(cmd *cobra.Command, args []string) error {
	languages := strings.Split(config.Render.Languages, ",")
	for _, v := range args {
		err := processFile(v, languages, config.Render.OutputDir, config.Render.LinkPrefix)
		if err != nil {
			return err
		}
	}
	return nil
}

func processFile(filePath string, types []string, outputDir string, linkPrefix string) error {
	err := validateFileExists(filePath)
	if err != nil {
		return err
	}

	// Read file into lines
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")

	// Construct a lookup for O(1) access
	typeLookup := make(map[string]bool)
	for _, v := range types {
		typeLookup[v] = true
	}

	// Find code blocks eligible for rendering
	var renderRegions []*Chunk
	for idx, line := range lines {
		if strings.HasPrefix(line, "```") {
			for k := range typeLookup {
				if strings.HasPrefix(line, fmt.Sprintf("```%s render", k)) {
					chunk, err := getRenderableChunk(lines, idx, k)
					if err != nil {
						return err
					}
					renderRegions = append(renderRegions, chunk)
					break
				}
			}
		}
	}

	// Construct a series of normal and render regions to represent the file
	var allChunks []*Chunk
	var currentIndex int
	for _, renderableChunk := range renderRegions {
		if currentIndex < renderableChunk.StartLineIndex {
			normalChunk := &Chunk{
				StartLineIndex: currentIndex,
				EndLineIndex:   renderableChunk.StartLineIndex - 1,
			}
			normalChunk.Lines = lines[normalChunk.StartLineIndex : normalChunk.EndLineIndex+1]
			allChunks = append(allChunks, normalChunk)
			allChunks = append(allChunks, renderableChunk)
			currentIndex = renderableChunk.EndLineIndex + 1
		}
	}
	if currentIndex < len(lines) {
		normalChunk := &Chunk{
			StartLineIndex: currentIndex,
			EndLineIndex:   len(lines) - 1,
		}
		normalChunk.Lines = lines[normalChunk.StartLineIndex : normalChunk.EndLineIndex+1]
		allChunks = append(allChunks, normalChunk)
	}

	// Render the renderable chunks and join the chunks back into a file
	var fileHasChanged bool
	var outputLines []string
	for _, chunk := range allChunks {
		if chunk.ShouldRender() {
			imageFileName, err := chunk.Render(outputDir, linkPrefix)
			if err != nil {
				return err
			}
			fmt.Printf("file=%s rendered=%s\n", filePath, imageFileName)
			fileHasChanged = true
		}
		outputLines = append(outputLines, chunk.Lines...)
	}

	// Write to disk
	if fileHasChanged {
		writer, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer writer.Close()
		output := strings.Join(outputLines, "\n")
		writer.WriteString(output)
	}

	return nil
}

func getRenderableChunk(lines []string, codeBlockIndex int, language string) (*Chunk, error) {
	chunk := &Chunk{}
	chunk.IsRenderable = true
	chunk.Language = language
	chunk.StartLineIndex = codeBlockIndex

	// Collect code block
	for i := codeBlockIndex + 1; i < len(lines); i++ {
		line := lines[i]
		if line == "```" {
			chunk.EndLineIndex = i
			break
		} else {
			chunk.CodeBlockContent = append(chunk.CodeBlockContent, line)
		}
	}

	// Check 2 lines above if the image has been rendered before
	for i := 1; i <= 2; i++ {
		idx := codeBlockIndex - i
		prevLine := lines[idx]
		matches := renderedImageRegexp.FindStringSubmatch(prevLine)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			chunk.StartLineIndex = idx
			chunk.ImageRelativeLineIndex = 0
			break
		}
	}

	chunk.Lines = append(chunk.Lines, lines[chunk.StartLineIndex:chunk.EndLineIndex+1]...)

	return chunk, nil
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

func buildMarkdownImage(outputFilename, linkPrefix string) string {
	return fmt.Sprintf("![%s](%s)", outputFilename, linkPrefix+outputFilename)
}
