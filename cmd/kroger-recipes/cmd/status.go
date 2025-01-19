/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/densestvoid/krogerrecipeshopper/migrations"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status of the server database schema",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("postgres", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s sslmode=disable",
			viper.GetString("db-host"),
			viper.GetInt("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-password"),
		))
		if err != nil {
			panic(err)
		}
		if err := migrations.Status(db); err != nil {
			panic(err)
		}
	},
}

func init() {
	migrateCmd.AddCommand(statusCmd)
}
