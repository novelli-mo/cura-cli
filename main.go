package main

import (
	"fmt"
	"os"

	"github.com/novelli-mo/cura/scanner"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "cura",
		Short:         "A skill manager for your repos",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize cura in the current repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := scanner.ScanRepo(dir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error scanning repo:", err)
				return err
			}

			fmt.Printf("Total files: %d\n", ctx.TotalFiles)
			fmt.Printf("Max depth: %d\n", ctx.MaxDepth)
			fmt.Printf("Root files: %v\n", ctx.RootFiles)
			fmt.Printf("Extensions: %v\n", ctx.Extensions)
			fmt.Printf("Folders: %v\n", ctx.FolderNames)
			if ctx.DocContent != "" {
				fmt.Printf("Docs found:\n%s\n", ctx.DocContent)
			}
			return nil
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
