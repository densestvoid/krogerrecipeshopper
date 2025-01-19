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
				"#ingredient-details-form",
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

func IngredientDetailsForm(ingredient *data.Ingredient) gomponents.Node {
	if ingredient != nil {
		return gomponents.Group{
			html.Input(
				html.Type("hidden"),
				html.Name("productID"),
				html.Value(ingredient.ProductID),
			),
			FormInput("ingredient-quantity", "ingredient-quantity", html.Input(
				html.ID("ingredient-quantity"),
				html.Class("form-control"),
				html.Type("number"),
				html.Name("quantity"),
				html.Min("0.01"),
				html.Step("0.5"),
				html.Value(fmt.Sprintf("%.2f", float32(ingredient.Quantity)/100)),
			)),
		}
	}
	return gomponents.Group{
		ProductsSearch(),
		FormInput("ingredient-quantity", "Ingredient quantity", html.Input(
			html.ID("ingredient-quantity"),
			html.Class("form-control"),
			html.Type("number"),
			html.Name("quantity"),
		)),
	}
}

func ProductsSearch() gomponents.Node {
	return html.Div(
		html.H3(gomponents.Text("Search products")),
		html.Input(
			html.Class("form-control"),
			html.Type("search"),
			html.Name("search"),
			html.Placeholder("Begin typing to seach products"),
			htmx.Post("/products/search"),
			htmx.Trigger("input changed delay:500ms, keyup[key=='Enter']"),
			htmx.Target("#products-search-table"),
			htmx.Indicator(".htmx-indicator"),
		),
		html.Span(html.Class("htmx-indicator"), gomponents.Text("Searching...")),
		html.Div(html.ID("products-search-table")),
	)
}

func ProductsSearchTable(products []Product) gomponents.Node {
	var productRows gomponents.Group
	for _, product := range products {
		productRows = append(productRows, ProductSearchRow(product))
	}
	return html.Table(
		html.THead(
			html.Tr(
				html.Th(gomponents.Text("Select")),
				html.Th(gomponents.Text("Image")),
				html.Th(gomponents.Text("Brand")),
				html.Th(gomponents.Text("Description")),
				html.Th(gomponents.Text("Size")),
			),
		),
		html.TBody(productRows),
	)
}

func ProductSearchRow(product Product) gomponents.Node {
	return html.Tr(
		html.Td(
			html.Input(
				html.Type("radio"),
				html.Name("productID"),
				html.Value(product.ProductID),
			),
		),
		html.Td(
			html.Img(
				html.Class("img-fluid img-thumbnail"),
				html.Src(product.ImageURL),
			),
		),
		html.Td(gomponents.Text(product.Brand)),
		html.Td(gomponents.Text(product.Description)),
		html.Td(gomponents.Text(product.Size)),
	)
}

type Product struct {
	ProductID   string
	Brand       string
	Description string
	Size        string
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
				html.Th(gomponents.Text("Size")),
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
		html.Td(gomponents.Text(ingredient.Size)),
		html.Td(gomponents.Text(fmt.Sprintf("%.2f", float32(ingredient.Quantity)/100))),
		html.Td(
			ModalButton(
				"ingredient-details-modal",
				"Edit details",
				fmt.Sprintf("/recipes/%v/ingredients/%s/details", ingredient.RecipeID, ingredient.ProductID),
				"#ingredient-details-form",
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
