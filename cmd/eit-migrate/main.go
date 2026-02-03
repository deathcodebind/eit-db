package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.5.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "eit-migrate",
		Short: "EIT Database Migration Tool",
		Long:  `A flexible migration tool for eit-db that supports both schema-based and raw SQL migrations.`,
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("eit-migrate version %s\n", version)
		},
	}
}
