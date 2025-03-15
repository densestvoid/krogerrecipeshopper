/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the kroger recipes server",
	Run: func(cmd *cobra.Command, args []string) {
		slog.SetLogLoggerLevel(slog.LevelDebug)

		db, err := sqlx.Open("pgx", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s sslmode=disable",
			viper.GetString("db-host"),
			viper.GetInt("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-password"),
		))
		if err != nil {
			panic(err)
		}
		repo := data.NewRepository(db)

		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.GetString("cache-host"), viper.GetInt("cache-port")),
			Password: viper.GetString("cache-password"),
		})
		if err := client.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}
		cache := data.NewCache(client, viper.GetDuration("cache-expiration"))

		handler := server.New(context.Background(), slog.Default(), server.Config{
			ClientID:     viper.GetString("client-id"),
			ClientSecret: viper.GetString("client-secret"),
			Domain:       viper.GetString("domain"),
		}, repo, cache)

		if !viper.GetBool("secure") {
			srv := http.Server{
				Addr:         fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port")),
				Handler:      handler,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
			}
			log.Fatal(srv.ListenAndServe())
		}

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS13,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		}

		srv := http.Server{
			Addr:         fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port")),
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		log.Fatal(srv.ListenAndServeTLS(viper.GetString("tls-cert"), viper.GetString("tls-key")))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server details
	serveCmd.Flags().StringP("host", "a", "localhost", "server host address")
	serveCmd.Flags().Int16P("port", "p", 8080, "server host port")
	serveCmd.Flags().BoolP("secure", "s", false, "secure connection (https)")
	serveCmd.Flags().String("tls-cert", "", "server certificate")
	serveCmd.Flags().String("tls-key", "", "server key")
	serveCmd.MarkFlagsRequiredTogether("secure", "tls-cert", "tls-key")

	// Cache details
	serveCmd.Flags().String("cache-host", "localhost", "cache host")
	serveCmd.Flags().Int("cache-port", 6379, "cache port")
	serveCmd.Flags().String("cache-password", "", "cache password")
	serveCmd.Flags().Duration("cache-expiration", time.Hour*24, "cache expiration")

	// Kroger application details
	serveCmd.Flags().String("client-id", "", "Kroger application id")
	serveCmd.Flags().String("client-secret", "", "Kroger application secret")
	serveCmd.Flags().String("domain", "", "Kroger apoplication domain for oath2 redirect url")

	// Bind all local flags to viper configuration variables
	if err := viper.BindPFlags(serveCmd.LocalFlags()); err != nil {
		panic(err)
	}
}
