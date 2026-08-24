package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

func NewTagsMux(repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			searchTerm := r.FormValue("search")
			tags := r.Form["tag"]

			tags, err := repo.ListTags(r.Context(), searchTerm, tags)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := templates.TagList(tags).Render(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
