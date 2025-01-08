package recipes

import (
	"github.com/go-chi/chi/v5"

	"github.com/densestvoid/krogerrecipeshopper/data"
)

func NewIngredientMux(repo *data.Repository) func(chi.Router) {
	return func(r chi.Router) {

	}
}
