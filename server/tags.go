package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates"
)

const MaxTagNameLength = 32

var ErrTagNameTooLong = errors.New("tag name exceeds maximum length")

func normalizeTagNames(tagNames []string) ([]string, error) {
	seen := make(map[string]struct{})
	filtered := []string{}
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		name = strings.ToLower(name)
		if len(name) > MaxTagNameLength {
			return nil, fmt.Errorf("%w (%d characters max): %q", ErrTagNameTooLong, MaxTagNameLength, name)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		filtered = append(filtered, name)
	}
	return filtered, nil
}

func NewTagsMux(repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			searchTerm := r.FormValue("search")
			excludedTags := r.Form["tag"]
			excludedTags, err := normalizeTagNames(excludedTags)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tags, err := repo.ListTags(r.Context(), searchTerm, excludedTags)
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
