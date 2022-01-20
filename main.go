package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

const (
	stateNone = iota
	stateInCodeBlock
)

// Match: ![render-db6d08bb022ed12c2cc74d86d7a4707d.svg](/optional/path/to/render-db6d08bb022ed12c2cc74d86d7a4707d.svg)
// Capture group on the hash.
var renderedImageRegexp = regexp.MustCompile(`!\[render-.{32}\..+\]\(.*render-(.{32})\..+\)`)

type Args struct {
	Types      []string // Languages to render
	OutputDir  string   // Directory to output rendered files to
	LinkPrefix string   // Prefix to use when linking to rendered files
	Files      []string // Markdown files to process
}

type Config struct {
	Render struct {
		OutputDir  string // Directory to output rendered files to
		Languages  string // Languages to render, comma separated
		LinkPrefix string // Prefix to use when linking to rendered files
	}
}

var config Config

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "md-code-render",
		Short: "A processor to render code blocks in Markdown files",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no files specified as input")
			}
			return nil
		},
		SilenceUsage: true,
	}

	cmd.AddCommand(NewRenderCmd())
	return cmd
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
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
