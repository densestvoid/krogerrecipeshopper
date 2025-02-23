package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/densestvoid/krogerrecipeshopper/assets"
	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Domain       string
}

func New(ctx context.Context, logger *slog.Logger, config Config, repo *data.Repository, cache *data.Cache) http.Handler {
	mux := chi.NewRouter()
	mux.Use(
		httprate.LimitByIP(100, time.Minute),
		NewErrorStatusMiddleware(),
		NewSlogMiddleware(),
	)

	// Handlers for favicons
	mux.Handle("/*", middleware.SetHeader("Cache-Control", "max-age=86400")(
		http.FileServer(http.FS(assets.Files)),
	))

	mux.Route("/auth", NewAuthMux(config, repo))
	mux.Group(func(r chi.Router) {
		r.Use(AuthenticationMiddleware(config, repo))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := templates.DashboardPage().Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		r.Route("/accounts", NewAccountMux(config, repo, cache))
		r.Route("/profiles", NewProfilesMux(repo))
		r.Route("/recipes", NewRecipesMux(config, repo, cache))
		r.Route("/products", NewProductsMux(config, repo, cache))
		r.Route("/locations", NewLocationsMux(config))
		r.Route("/cart", NewCartMux(config, repo, cache))
	})

	return mux
}
