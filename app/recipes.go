package app

import (
	"github.com/google/uuid"

	"github.com/densestvoid/krogerrecipeshopper/data"
)

func CanViewRecipe(recipe data.Recipe, viewerAccountID uuid.UUID) bool {
	if recipe.AccountID == viewerAccountID {
		return true
	}
	return recipe.Visibility == data.VisibilityPublic
}
