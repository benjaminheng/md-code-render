package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var renderedImageFilenameRegexp = regexp.MustCompile(`render-.{32}\.(svg|png)`)

func NewCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove orphaned images not linked to in any Markdown file",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no files specified as input")
			}
			return nil
		},
		RunE: cleanCmd,
	}
	cmd.Flags().StringVar(&config.Clean.ImageDir, "image-dir", "", "(required) Directory containing images")
	cmd.MarkFlagRequired("image-dir")
	return cmd
}

func cleanCmd(cmd *cobra.Command, args []string) error {
	// Collect all file contents
	var allContent string
	for _, v := range args {
		b, err := os.ReadFile(v)
		if err != nil {
			return err
		}
		allContent += "\n" + string(b)
	}

	// Collect files to remove
	var filesToRemove []string
	entries, err := os.ReadDir(config.Clean.ImageDir)
	if err != nil {
		return err
	}
	for _, v := range entries {
		if v.IsDir() {
			continue
		}
		if !renderedImageFilenameRegexp.MatchString(v.Name()) {
			continue
		}
		// This is not efficient, since we are iterating through the
		// contents of all files for each image being checked.
		// Candidate for optimization later.
		if !strings.Contains(allContent, v.Name()) {
			filesToRemove = append(filesToRemove, path.Join(config.Clean.ImageDir, v.Name()))
		}
	}

	// Remove files
	for _, v := range filesToRemove {
		err := os.Remove(v)
		if err != nil {
			return err
		}
		fmt.Printf("Removed orphaned file %s\n", v)
	}
	return nil
}
