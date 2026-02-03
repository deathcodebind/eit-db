package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func upCmd() *cobra.Command {
	var migrationDir string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		Long:  `Executes all migrations that haven't been applied yet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrationCommand(migrationDir, "up")
		},
	}

	cmd.Flags().StringVarP(&migrationDir, "dir", "d", "migrations", "Directory containing migrations")

	return cmd
}

func downCmd() *cobra.Command {
	var migrationDir string

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		Long:  `Rolls back the most recently applied migration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrationCommand(migrationDir, "down")
		},
	}

	cmd.Flags().StringVarP(&migrationDir, "dir", "d", "migrations", "Directory containing migrations")

	return cmd
}

func statusCmd() *cobra.Command {
	var migrationDir string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Displays the status of all migrations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrationCommand(migrationDir, "status")
		},
	}

	cmd.Flags().StringVarP(&migrationDir, "dir", "d", "migrations", "Directory containing migrations")

	return cmd
}

func runMigrationCommand(migrationDir, command string) error {
	fmt.Printf("Running migrations from %s...\n", migrationDir)
	fmt.Printf("\nPlease run the following command:\n")
	fmt.Printf("  cd %s && go run . %s\n", migrationDir, command)
	fmt.Printf("\nNote: Make sure you have configured your database credentials in %s/.env\n", migrationDir)
	
	return nil
}
