package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
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

var defaultRenderOptions = RenderOptions{Mode: "normal"}

type RenderOptions struct {
	Mode string `json:"mode"` // Modes: normal, code-collapsed, image-collapsed, code-hidden
}

func (o *RenderOptions) Validate() error {
	switch o.Mode {
	case "normal", "code-collapsed", "image-collapsed", "code-hidden":
	default:
		return errors.New("unsupported mode")
	}
	return nil
}

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
	RenderOptions          RenderOptions
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
		ext = "svg"
		content, err = runShellCommand("dot", []string{"-Tsvg"}, strings.NewReader(strings.Join(r.CodeBlockContent, "\n")))
		if err != nil {
			return "", err
		}
	case "plantuml":
		ext = "svg"
		content, err = runShellCommand("plantuml", []string{"-tsvg", "-pipe"}, strings.NewReader(strings.Join(r.CodeBlockContent, "\n")))
		if err != nil {
			return "", err
		}
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
	r.Lines[r.ImageRelativeLineIndex] = image

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
	cmd.Flags().StringVar(&config.Render.Languages, "languages", "", "(required) Languages to render. Comma-separated. Supported languages: [dot, plantuml].")
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

	// Split the file into chunks. A chunk can represent either a normal
	// segment, or a renderable segment.
	var chunks []*Chunk
	var lastChunkIndex int
	for idx, line := range lines {
		// Skip ahead if these lines have been assigned a chunk already
		if idx < lastChunkIndex {
			continue
		}
		// Look for renderable code blocks
		if strings.HasPrefix(line, "```") {
			for k := range typeLookup {
				if strings.HasPrefix(line, fmt.Sprintf("```%s render", k)) {
					// Look at lines in and around the code
					// block to determine the renderable chunk.
					renderChunk, err := getRenderableChunk(lines, idx, k)
					if err != nil {
						return err
					}
					// Preceding lines not part of the renderable chunk are part of a
					// normal chunk; construct one and add it to our list of chunks.
					normalChunk := &Chunk{
						StartLineIndex: lastChunkIndex,
						EndLineIndex:   renderChunk.StartLineIndex - 1,
					}
					normalChunk.Lines = lines[normalChunk.StartLineIndex : normalChunk.EndLineIndex+1]
					chunks = append(chunks, normalChunk, renderChunk)
					lastChunkIndex = renderChunk.EndLineIndex + 1
					break
				}
			}
		}
	}
	if lastChunkIndex < len(lines) {
		// The rest of the file is a normal chunk
		normalChunk := &Chunk{
			StartLineIndex: lastChunkIndex,
			EndLineIndex:   len(lines) - 1,
		}
		normalChunk.Lines = lines[normalChunk.StartLineIndex : normalChunk.EndLineIndex+1]
		chunks = append(chunks, normalChunk)
	}

	// Render the renderable chunks and join the chunks back into a file
	var fileHasChanged bool
	var outputLines []string
	for _, chunk := range chunks {
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

	fence := lines[codeBlockIndex]
	renderOptionsJSON := strings.TrimPrefix(fence, fmt.Sprintf("```%s render", language))
	if strings.HasPrefix(renderOptionsJSON, "{") && strings.HasSuffix(renderOptionsJSON, "}") {
		var renderOptions RenderOptions
		err := json.Unmarshal([]byte(renderOptionsJSON), &renderOptions)
		if err != nil {
			return nil, err
		}
		err = renderOptions.Validate()
		if err != nil {
			return nil, err
		}
		chunk.RenderOptions = renderOptions
	} else {
		chunk.RenderOptions = defaultRenderOptions
	}

	var err error
	renderTemplateManager := RenderTemplateManager{}
	switch chunk.RenderOptions.Mode {
	case "normal":
		err = renderTemplateManager.Normal(lines, codeBlockIndex, chunk)
	case "code-collapsed":
		err = renderTemplateManager.CodeCollapsed(lines, codeBlockIndex, chunk)
	case "image-collapsed":
		err = renderTemplateManager.ImageCollapsed(lines, codeBlockIndex, chunk)
	case "code-hidden":
		err = renderTemplateManager.CodeHidden(lines, codeBlockIndex, chunk)
	default:
		return nil, errors.New("unsupported mode")
	}
	if err != nil {
		return nil, err
	}

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

// RenderTemplateManager contains methods to handle the templates for different rendering modes.
type RenderTemplateManager struct{}

// Normal handles the template for the "normal" mode. The template looks like:
//
//	![]()
//
//	```dot render
//	```
func (m RenderTemplateManager) Normal(lines []string, codeBlockIndex int, chunk *Chunk) (err error) {
	content, codeBlockEndIndex, fenceStart, fenceEnd, err := m.collectCodeBlock(lines, codeBlockIndex)
	if err != nil {
		return err
	}
	chunk.CodeBlockContent = content
	chunk.StartLineIndex = codeBlockIndex
	chunk.EndLineIndex = codeBlockEndIndex

	var isRenderedBefore bool
	// Check 2 lines above if the image has been rendered before
	for i := 1; i <= 2; i++ {
		idx := codeBlockIndex - i
		prevLine := lines[idx]
		matches := renderedImageRegexp.FindStringSubmatch(prevLine)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			chunk.StartLineIndex = idx
			chunk.ImageRelativeLineIndex = 0
			isRenderedBefore = true
			break
		}
	}

	// Render the template into the chunk. Image will be replaced later.
	if !isRenderedBefore {
		chunk.Lines = []string{"<!-- image here -->", "", fenceStart}
		chunk.Lines = append(chunk.Lines, chunk.CodeBlockContent...)
		chunk.Lines = append(chunk.Lines, fenceEnd)
		chunk.ImageRelativeLineIndex = 0
		chunk.RenderedHash = ""
	} else {
		chunk.Lines = lines[chunk.StartLineIndex : chunk.EndLineIndex+1]
	}
	return nil
}

// CodeCollapsed handles the template for the "code-collapsed" mode. The template looks like:
//
//	![]()
//
//	<details><summary>Source</summary>
//
//	```dot render
//	```
//
//	</details>
func (m RenderTemplateManager) CodeCollapsed(lines []string, codeBlockIndex int, chunk *Chunk) (err error) {
	content, codeBlockEndIndex, fenceStart, fenceEnd, err := m.collectCodeBlock(lines, codeBlockIndex)
	if err != nil {
		return err
	}
	chunk.CodeBlockContent = content
	chunk.StartLineIndex = codeBlockIndex
	chunk.EndLineIndex = codeBlockEndIndex

	// Check if rendered before
	closingDetailsTag := "</details>"
	hasClosingDetailsTag := codeBlockEndIndex+2 < len(lines) && lines[codeBlockEndIndex+2] == closingDetailsTag
	openingDetailsTag := "<details><summary>Source</summary>"
	hasOpeningDetailsTag := codeBlockIndex-2 >= 0 && lines[codeBlockIndex-2] == openingDetailsTag
	var hasImage bool
	if codeBlockIndex-4 >= 0 {
		line := lines[codeBlockIndex-4]
		matches := renderedImageRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			chunk.StartLineIndex = codeBlockIndex - 4
			chunk.ImageRelativeLineIndex = 0
			hasImage = true
		}
	}

	// Render the template into the chunk. Image will be replaced later.
	isRenderedBefore := hasClosingDetailsTag && hasOpeningDetailsTag && hasImage
	if !isRenderedBefore {
		chunk.Lines = []string{"<!-- image here -->", "", openingDetailsTag, "", fenceStart}
		chunk.Lines = append(chunk.Lines, chunk.CodeBlockContent...)
		chunk.Lines = append(chunk.Lines, fenceEnd, "", closingDetailsTag)
		chunk.ImageRelativeLineIndex = 0
		chunk.RenderedHash = ""
	} else {
		chunk.Lines = lines[chunk.StartLineIndex : chunk.EndLineIndex+1]
	}
	return nil
}

// ImageCollapsed handles the template for the "image-collapsed" mode. The template looks like:
//
//	```dot render
//	```
//
//	<details><summary>Image</summary>
//
//	![]()
//
//	</details>
func (m RenderTemplateManager) ImageCollapsed(lines []string, codeBlockIndex int, chunk *Chunk) (err error) {
	content, codeBlockEndIndex, fenceStart, fenceEnd, err := m.collectCodeBlock(lines, codeBlockIndex)
	if err != nil {
		return err
	}
	chunk.CodeBlockContent = content
	chunk.StartLineIndex = codeBlockIndex
	chunk.EndLineIndex = codeBlockEndIndex

	// Check if rendered before
	openingDetailsTag := "<details><summary>Image</summary>"
	hasOpeningDetailsTag := codeBlockEndIndex+2 < len(lines) && lines[codeBlockEndIndex+2] == openingDetailsTag
	closingDetailsTag := "</details>"
	hasClosingDetailsTag := codeBlockEndIndex+6 < len(lines) && lines[codeBlockEndIndex+6] == closingDetailsTag
	var hasImage bool
	if codeBlockEndIndex+4 < len(lines) {
		line := lines[codeBlockEndIndex+4]
		matches := renderedImageRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			chunk.EndLineIndex = codeBlockEndIndex + 6
			chunk.ImageRelativeLineIndex = (chunk.EndLineIndex - chunk.StartLineIndex) - 2
			hasImage = true
		}
	}

	// Render the template into the chunk. Image will be replaced later.
	isRenderedBefore := hasClosingDetailsTag && hasOpeningDetailsTag && hasImage
	if !isRenderedBefore {
		chunk.Lines = append([]string{fenceStart}, chunk.CodeBlockContent...)
		chunk.Lines = append(chunk.Lines, []string{fenceEnd, "", openingDetailsTag, "", "<!-- image here --", "", closingDetailsTag}...)
		chunk.ImageRelativeLineIndex = len(chunk.Lines) - 3
		chunk.RenderedHash = ""
	} else {
		chunk.Lines = lines[chunk.StartLineIndex : chunk.EndLineIndex+1]
	}
	return nil
}

// CodeHidden handles the template for the "code-hidden" mode. The template looks like:
//
//	![]()
//
//	<!--
//	```dot render
//	```
//	-->
func (m RenderTemplateManager) CodeHidden(lines []string, codeBlockIndex int, chunk *Chunk) (err error) {
	content, codeBlockEndIndex, fenceStart, fenceEnd, err := m.collectCodeBlock(lines, codeBlockIndex)
	if err != nil {
		return err
	}
	chunk.CodeBlockContent = content
	chunk.StartLineIndex = codeBlockIndex
	chunk.EndLineIndex = codeBlockEndIndex

	// Check if rendered before
	openingCommentTag := "<!--"
	hasOpeningCommentTag := codeBlockIndex-1 > 0 && lines[codeBlockIndex-1] == openingCommentTag
	closingCommentTag := "-->"
	hasClosingCommentTag := codeBlockEndIndex+1 < len(lines) && lines[codeBlockEndIndex+1] == closingCommentTag
	var hasImage bool
	if codeBlockIndex-3 > 0 {
		line := lines[codeBlockIndex-3]
		matches := renderedImageRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			chunk.StartLineIndex = codeBlockIndex - 3
			chunk.ImageRelativeLineIndex = 0
			hasImage = true
		}
	}

	// Render the template into the chunk. Image will be replaced later.
	isRenderedBefore := hasOpeningCommentTag && hasClosingCommentTag && hasImage
	if !isRenderedBefore {
		chunk.Lines = []string{"<!-- image here -->", "", openingCommentTag, fenceStart}
		chunk.Lines = append(chunk.Lines, chunk.CodeBlockContent...)
		chunk.Lines = append(chunk.Lines, fenceEnd, closingCommentTag)
		chunk.ImageRelativeLineIndex = 0
		chunk.RenderedHash = ""
	} else {
		chunk.Lines = lines[chunk.StartLineIndex : chunk.EndLineIndex+1]
	}
	return nil
}

func (m RenderTemplateManager) collectCodeBlock(lines []string, codeBlockIndex int) (content []string, codeBlockEndIndex int, fenceStart string, fenceEnd string, err error) {
	for i := codeBlockIndex + 1; i < len(lines); i++ {
		line := lines[i]
		if line == "```" {
			return content, i, lines[codeBlockIndex], line, nil
		}
		content = append(content, line)
	}
	return nil, 0, "", "", errors.New("code block is unterminated")
}
