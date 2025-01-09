package templates

import (
	"fmt"

	"github.com/google/uuid"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
)

func Ingredients(recipe *data.Recipe) gomponents.Node {
	return BasePage("Ingredients", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text(fmt.Sprintf("%s ingredients", recipe.Name)),
			),
			html.Div(
				htmx.Get(fmt.Sprintf("/recipes/%v/ingredients/table", recipe.ID)),
				htmx.Swap("innerHTML"),
				htmx.Trigger("load,ingredient-update from:body"),
			),
			ModalButton(
				"ingredient-details-modal",
				"Add ingredient",
				fmt.Sprintf("/recipes/%v/ingredients//details", recipe.ID),
				"#ingredients-details-form",
			),
		),
		Modal("ingredient-details-modal", "Edit ingredient",
			gomponents.Group{},
			gomponents.Group{
				html.Form(
					html.ID("ingredient-details-form"),
				),
			},
			gomponents.Group{
				html.Button(
					html.Type("submit"),
					html.Class("btn btn-primary"),
					html.Data("bs-dismiss", "modal"),
					gomponents.Text("Submit"),
					htmx.Include("#ingredient-details-form"),
					htmx.Post(fmt.Sprintf("/recipes/%v/ingredients", recipe.ID)),
					htmx.Swap("none"),
				),
			},
		),
	})
}

func IngredientsSearch() gomponents.Node {
	return html.Div(
		html.H3(gomponents.Text("Search ingredients")),
		html.Input(
			html.Class("form-control"),
			html.Type("search"),
			html.Name("search"),
			html.Placeholder("Begin typing to seach ingredients"),
			htmx.Get("/ingredients/search"),
			htmx.Target("#ingredients-table"),
			htmx.Indicator(".htmx-indicator"),
		),
	)
}

type Product struct {
	ProductID   string
	Brand       string
	Description string
	ImageURL    string
}

type Ingredient struct {
	Product
	RecipeID uuid.UUID
	Quantity int
}

func IngredientsTable(ingredients []Ingredient) gomponents.Node {
	var ingredientRows gomponents.Group
	for _, ingredient := range ingredients {
		ingredientRows = append(ingredientRows, IngredientRow(ingredient))
	}
	return html.Table(
		html.Class("table table-striped table-bordered text-center align-middle w-100"),
		html.THead(
			html.Tr(
				html.Th(gomponents.Text("Image")),
				html.Th(gomponents.Text("Brand")),
				html.Th(gomponents.Text("Description")),
				html.Th(gomponents.Text("Quantity")),
				html.Th(gomponents.Text("Actions")),
			),
		),
		html.TBody(
			html.Class("table-group-divider"),
			ingredientRows,
		),
	)
}

func IngredientRow(ingredient Ingredient) gomponents.Node {
	return html.Tr(
		html.Td(
			html.Img(
				html.Class("img-fluid img-thumbnail"),
				html.Src(ingredient.ImageURL),
			),
		),
		html.Td(gomponents.Text(ingredient.Brand)),
		html.Td(gomponents.Text(ingredient.Description)),
		html.Td(gomponents.Text(fmt.Sprintf("%.2f", float32(ingredient.Quantity)/100))),
		html.Td(
			ModalButton(
				"ingredients-details-modal",
				"Edit details",
				fmt.Sprintf("/recipes/%v/ingredients/%s", ingredient.RecipeID, ingredient.ProductID),
				"#ingredients-details-form",
			),
			html.Button(
				html.Type("button"),
				html.Class("btn btn-danger"),
				gomponents.Text("Delete"),
				htmx.Delete(fmt.Sprintf("/recipes/%v/ingredients/%s", ingredient.RecipeID, ingredient.ProductID)),
				htmx.Swap("none"),
			),
		),
	)
}
