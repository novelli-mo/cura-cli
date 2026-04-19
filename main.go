package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cura",
		Short: "A skill manager for your repos",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize cura in the current repo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("initializing...")
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
