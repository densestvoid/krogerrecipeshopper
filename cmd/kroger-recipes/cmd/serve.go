/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
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
		repo := data.NewRepository(db)

		handler := server.New(context.Background(), slog.Default(), server.Config{
			ClientID:     viper.GetString("client-id"),
			ClientSecret: viper.GetString("client-secret"),
			Domain:       viper.GetString("domain"),
		}, repo)

		if !viper.GetBool("secure") {
			srv := http.Server{
				Addr:    fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port")),
				Handler: handler,
			}
			log.Fatal(srv.ListenAndServe())
		}

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		srv := http.Server{
			Addr:         fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port")),
			Handler:      handler,
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

	// Kroger application details
	serveCmd.Flags().String("client-id", "", "Kroger application id")
	serveCmd.Flags().String("client-secret", "", "Kroger application secret")
	serveCmd.Flags().String("domain", "", "Kroger apoplication domain for oath2 redirect url")

	// Bind all local flags to viper configuration variables
	if err := viper.BindPFlags(serveCmd.LocalFlags()); err != nil {
		panic(err)
	}
}
