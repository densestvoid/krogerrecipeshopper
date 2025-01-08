package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	recipes "github.com/densestvoid/krogerrecipeshopper"
	"github.com/densestvoid/krogerrecipeshopper/data"
)

const (
	ClientID      = "recipe-shopper-276be53a09a3bb09150aef03b5783ebc7840425982417759754"
	ClientSecret  = "L6jkJmLdTBysOzNFABmKE06qa-5qwUW_tEXga1g-"
	OAuth2BaseURL = "https://api.kroger.com/v1/connect/oauth2"
	RedirectUrl   = "https://4a05-50-5-199-176.ngrok-free.app/auth"

	DatabaseAddr     = "localhost"
	DatabasePort     = "32770"
	DatabaseUser     = "kroger"
	DatabasePassword = "password"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		DatabaseAddr,
		DatabasePort,
		DatabaseUser,
		DatabasePassword,
	))
	if err != nil {
		panic(err)
	}
	repo := data.NewRepository(db)

	handler := recipes.NewServer(context.Background(), slog.Default(), recipes.Config{
		ClientID:      ClientID,
		ClientSecret:  ClientSecret,
		OAuth2BaseURL: OAuth2BaseURL,
		RedirectUrl:   RedirectUrl,
	}, repo)

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  time.Second * 500,
		WriteTimeout: time.Second * 500,
		IdleTimeout:  time.Second * 500,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
