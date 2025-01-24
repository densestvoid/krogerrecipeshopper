package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Domain       string
}

func New(ctx context.Context, logger *slog.Logger, config Config, repo *data.Repository) http.Handler {
	mux := chi.NewRouter()
	chiLogger := httplog.NewLogger("Kroger Recipes", httplog.Options{
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		RequestHeaders:  true,
		ResponseHeaders: true,
	})
	mux.Use(httplog.RequestLogger(chiLogger))

	mux.Route("/auth", NewAuthMux(config, repo))
	mux.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware(config, repo))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.DashboardPage().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		r.Route("/accounts", NewAccountMux(repo))
		r.Route("/profiles", NewProfilesMux(repo))
		r.Route("/recipes", NewRecipesMux(config, repo))
		r.Route("/products", NewProductsMux(config, repo))
		r.Route("/cart", NewCartMux(repo, config))
	})

	return mux
}

func Error(w http.ResponseWriter, message string, err error, statusCode int) {
	slog.Error(message, "error", err)
	if err := templates.Alert("Failed", "alert-danger").Render(w); err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}
