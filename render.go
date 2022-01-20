package main

import "github.com/spf13/cobra"

func NewRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render code blocks in markdown files",
		Long:  ``,
		RunE:  renderCmd,
	}
	cmd.Flags().StringVar(&config.Render.OutputDir, "output-dir", "", "Directory to render code blocks to. If not specified, output will be rendered to the same directory as the input file.")
	cmd.Flags().StringVar(&config.Render.Languages, "languages", "", "Languages to render (comma-separated, supported languages: [dot])")
	cmd.MarkFlagRequired("languages")
	cmd.Flags().StringVar(&config.Render.LinkPrefix, "link-prefix", "", "Prefix to use when linking to rendered files")
	return cmd
}

func renderCmd(cmd *cobra.Command, args []string) error {
	return nil
}
