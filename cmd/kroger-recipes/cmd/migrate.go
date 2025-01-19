/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the server database schema",
	Long: `Migrate the server database schema up and down
	to support different versions of the server.`,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
