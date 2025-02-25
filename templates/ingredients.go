package templates

import (
	"fmt"

	"github.com/google/uuid"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
)

func Ingredients(accountID uuid.UUID, recipe data.Recipe) gomponents.Node {
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
			gomponents.If(accountID == recipe.AccountID, ModalButton(
				"btn-primary",
				"Add ingredient",
				htmx.Get(fmt.Sprintf("/recipes/%v/ingredients//details", recipe.ID)),
			)),
		),
	})
}

func IngredientDetailsModalContent(recipeID uuid.UUID, ingredient data.Ingredient) gomponents.Node {
	ifExists := func(node gomponents.Node) gomponents.Node {
		return gomponents.If(ingredient.ProductID != "", node)
	}

	ifNotExists := func(node gomponents.Node) gomponents.Node {
		return gomponents.If(ingredient.ProductID == "", node)
	}

	return ModalContent(
		"Ingredient Details",
		ModalForm(
			htmx.Post(fmt.Sprintf("/recipes/%v/ingredients", recipeID)),
			ifExists(html.Input(
				html.Type("hidden"),
				html.Name("productID"),
				html.Value(ingredient.ProductID),
			)),
			ifNotExists(ProductsSearch()),
			html.Div(
				gomponents.Attr("x-data", fmt.Sprintf("{staple : %t}", ingredient.Staple)),
				html.Class("input-group"),
				html.Div(
					html.Class("input-group-text"),
					FormCheck("ingredient-staple", "Staple", true, html.Input(
						gomponents.Attr("x-on:change", "staple = !staple"),
						html.ID("ingredient-staple"),
						html.Class("form-check-input"),
						html.Name("staple"),
						html.Role("switch"),
						html.Type("checkbox"),
						html.Value("true"),
						Checked(ingredient.Staple),
					)),
				),
				FormInput("ingredient-quantity", "Ingredient quantity", nil,
					html.Input(
						gomponents.Attr("x-bind:value", fmt.Sprintf("staple ? '1' : '%s'", ingredient.QuantityDecimalString())),
						gomponents.Attr("x-bind:disabled", "staple"),
						html.Class("form-control"),
						html.Type("number"),
						html.Name("quantity"),
						html.Min("0.01"),
						html.Step("0.01"),
						html.Required(),
						ifExists(html.Value(ingredient.QuantityDecimalString())),
					),
				),
			),
		),
		gomponents.Group{
			ModalDismiss(),
			ModalSubmit(),
		},
	)
}

func Checked(b bool) gomponents.Node {
	if b {
		return html.Checked()
	}
	return nil
}

func ProductsSearch() gomponents.Node {
	return html.Div(
		html.H3(gomponents.Text("Search products")),
		html.Div(
			html.Form(
				html.Input(
					html.Class("form-control"),
					html.Type("search"),
					html.Name("search"),
					html.Placeholder("Begin typing to seach products"),
					htmx.Post("/products/search"),
					htmx.Swap("innerHTML"),
					htmx.Trigger("input changed delay:1s, keyup[key=='Enter']"),
					htmx.Target("#products-search-table"),
					htmx.Indicator(".htmx-indicator"),
				),
			),
		),
		html.Span(html.Class("htmx-indicator"), gomponents.Text("Searching...")),
		html.Input(
			html.Class("d-none"),
			html.Type("radio"),
			html.Name("productID"),
			html.Required(),
		),
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
		html.Td(html.A(
			html.Href(product.ProductURL),
			html.Target("_blank"),
			gomponents.Text(product.Description),
		)),
		html.Td(gomponents.Text(product.Size)),
	)
}

type Product struct {
	ProductID   string
	Brand       string
	Description string
	Size        string
	ImageURL    string
	ProductURL  string
}

type Ingredient struct {
	Product
	RecipeID uuid.UUID
	Quantity int
	Staple   bool
}

func IngredientsTable(accountID, recipeAccountID uuid.UUID, ingredients []Ingredient) gomponents.Node {
	var ingredientRows, stapleRows gomponents.Group
	for _, ingredient := range ingredients {
		if ingredient.Staple {
			stapleRows = append(stapleRows, IngredientRow(accountID, recipeAccountID, ingredient))
		} else {
			ingredientRows = append(ingredientRows, IngredientRow(accountID, recipeAccountID, ingredient))
		}
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
				gomponents.If(accountID == recipeAccountID, html.Th(gomponents.Text("Actions"))),
			),
		),
		html.TBody(
			html.Class("table-group-divider"),
			html.Tr(html.Td(html.ColSpan("6"), gomponents.Text("Ingredients"))),
			ingredientRows,
		),
		html.TBody(
			html.Class("table-group-divider"),
			html.Tr(html.Td(html.ColSpan("6"), gomponents.Text("Staples"))),
			stapleRows,
		),
	)
}

func IngredientRow(accountID, recipeAccountID uuid.UUID, ingredient Ingredient) gomponents.Node {
	return html.Tr(
		html.Td(
			html.Img(
				html.Class("img-fluid img-thumbnail"),
				html.Src(ingredient.ImageURL),
			),
		),
		html.Td(gomponents.Text(ingredient.Brand)),
		html.Td(html.A(
			html.Href(ingredient.ProductURL),
			html.Target("_blank"),
			gomponents.Text(ingredient.Description),
		)),
		html.Td(gomponents.Text(ingredient.Size)),
		html.Td(gomponents.Textf("%.2f", float64(ingredient.Quantity)/100)),
		gomponents.If(accountID == recipeAccountID, html.Td(
			ModalButton(
				"btn-primary",
				"Edit details",
				htmx.Get(fmt.Sprintf("/recipes/%v/ingredients/%s/details", ingredient.RecipeID, ingredient.ProductID)),
			),
			html.Button(
				html.Type("button"),
				html.Class("btn btn-danger"),
				gomponents.Text("Delete"),
				htmx.Delete(fmt.Sprintf("/recipes/%v/ingredients/%s", ingredient.RecipeID, ingredient.ProductID)),
				htmx.Swap("none"),
				htmx.Confirm("Are you sure you want to remove this ingredient from the recipe?"),
			),
		)),
	)
}
