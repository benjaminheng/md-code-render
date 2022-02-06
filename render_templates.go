package main

import "errors"

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
		hasImage := m.checkForImage(chunk, prevLine, func() {
			chunk.StartLineIndex = idx
			chunk.ImageRelativeLineIndex = 0
			isRenderedBefore = true
		})
		if hasImage {
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
		hasImage = m.checkForImage(chunk, line, func() {
			chunk.StartLineIndex = codeBlockIndex - 4
			chunk.ImageRelativeLineIndex = 0
		})
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
		hasImage = m.checkForImage(chunk, line, func() {
			chunk.EndLineIndex = codeBlockEndIndex + 6
			chunk.ImageRelativeLineIndex = (chunk.EndLineIndex - chunk.StartLineIndex) - 2
		})
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
		hasImage = m.checkForImage(chunk, line, func() {
			chunk.StartLineIndex = codeBlockIndex - 3
			chunk.ImageRelativeLineIndex = 0
		})
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

func (m RenderTemplateManager) checkForImage(chunk *Chunk, line string, imageExistsFn func()) (imageExists bool) {
	if chunk.RenderOptions.Filename != "" {
		matches := markdownImageRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			imageExistsFn()
			return true
		}
	} else {
		matches := renderedImageRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			chunk.RenderedHash = matches[1]
			imageExistsFn()
			return true
		}
	}
	return false
}
