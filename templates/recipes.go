package templates

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
)

func Recipes() gomponents.Node {
	return BasePage("Recipes", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Recipes"),
			),
			html.Div(
				htmx.Get("/recipes/table"),
				htmx.Swap("innerHTML"),
				htmx.Trigger("load,recipe-update from:body"),
			),
			ModalButton(
				"recipe-details-modal",
				"Add recipe",
				"",
				"/recipes//details",
				"#recipe-details-form",
			),
		),
		Modal("recipe-details-modal", "Edit recipe",
			gomponents.Group{},
			gomponents.Group{
				html.Form(
					html.ID("recipe-details-form"),
				),
			},
			gomponents.Group{
				html.Button(
					html.Type("submit"),
					html.Class("btn btn-primary"),
					html.Data("bs-dismiss", "modal"),
					gomponents.Text("Submit"),
					htmx.Include("#recipe-details-form"),
					htmx.Post("/recipes"),
					htmx.Swap("none"),
				),
			},
		),
	})
}

func RecipeDetailsForm(recipe *data.Recipe) gomponents.Node {
	if recipe != nil {
		return gomponents.Group{
			html.Input(
				html.Type("hidden"),
				html.Name("id"),
				html.Value(recipe.ID.String()),
			),
			FormInput("recipe-name", "Recipe name", html.Input(
				html.ID("recipe-name"),
				html.Class("form-control"),
				html.Type("text"),
				html.Name("name"),
				html.Value(recipe.Name),
			)),
			FormInput("recipe-description", "Recipe description", html.Input(
				html.ID("recipe-description"),
				html.Class("form-control"),
				html.Type("text"),
				html.Name("description"),
				html.Value(recipe.Description),
			)),
		}
	}
	return gomponents.Group{
		FormInput("recipe-name", "Recipe name", html.Input(
			html.ID("recipe-name"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("name"),
		)),
		FormInput("recipe-description", "Recipe description", html.Input(
			html.ID("recipe-description"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("description"),
		)),
	}
}

func FormInput(id, label string, input gomponents.Node) gomponents.Node {
	return html.Div(
		html.Class("form-floating"),
		input,
		html.Label(
			html.For(id),
			gomponents.Text(label),
		),
	)
}

func RecipeTable(recipes []data.Recipe) gomponents.Node {
	var recipeRows gomponents.Group
	for _, recipe := range recipes {
		recipeRows = append(recipeRows, RecipeRow(recipe))
	}
	return html.Table(
		html.Class("table table-striped table-bordered text-center align-middle w-100"),
		html.THead(
			html.Tr(
				html.Th(gomponents.Text("Name")),
				html.Th(
					// Hide if the screen is small
					html.Class("d-none d-sm-table-cell"),
					gomponents.Text("Description"),
				),
				html.Th(gomponents.Text("Actions")),
			),
		),
		html.TBody(
			html.Class("table-group-divider"),
			recipeRows,
		),
	)
}

func RecipeRow(recipe data.Recipe) gomponents.Node {
	return html.Tr(
		html.Td(gomponents.Text(recipe.Name)),
		html.Td(
			// Hide if the screen is small
			html.Class("d-none d-sm-table-cell"),
			gomponents.Text(recipe.Description)),
		html.Td(
			html.Div(
				html.Class("btn-group"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary"),
					gomponents.Text("Add to cart"),
					htmx.Post(fmt.Sprintf("/cart/recipe/%v", recipe.ID)),
					htmx.Swap("none"),
				),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary dropdown-toggle dropdown-toggle-split"),
					html.Data("bs-toggle", "dropdown"),
				),
				html.Ul(
					html.Class("dropdown-menu"),
					html.Li(
						ModalButton(
							"recipe-details-modal",
							"Edit details",
							"w-100",
							fmt.Sprintf("/recipes/%s/details", recipe.ID.String()),
							"#recipe-details-form",
						),
					),
					html.Li(
						html.A(
							html.Href(fmt.Sprintf("/recipes/%v/ingredients", recipe.ID)),
							html.Button(
								html.Type("button"),
								html.Class("btn btn-secondary w-100"),
								gomponents.Text("Edit ingredients"),
							),
						),
					),
					html.Li(html.Hr(html.Class("dropdown-divider"))),
					html.Li(
						html.Button(
							html.Type("button"),
							html.Class("btn btn-danger w-100"),
							gomponents.Text("Delete"),
							htmx.Delete(fmt.Sprintf("/recipes/%v", recipe.ID)),
							htmx.Swap("none"),
							htmx.Confirm("Are you sure you want to delete thi recipe?"),
						),
					),
				),
			),
		),
	)
}
