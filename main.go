package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Config struct {
	Clean struct {
		ImageDir string
	}
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
		SilenceUsage: true,
	}

	cmd.AddCommand(NewRenderCmd())
	cmd.AddCommand(NewCleanCmd())
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
