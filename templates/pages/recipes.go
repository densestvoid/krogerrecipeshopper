package pages

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/densestvoid/krogerrecipeshopper/templates/components"
)

func Recipes() gomponents.Node {
	return BasePage("Recipes", "/", gomponents.Group{
		html.Div(
			htmx.Get("/recipes/table"),
			htmx.Swap("innerHTML"),
			htmx.Trigger("load,recipe-update from:body"),
		),
		components.ModalButton(
			"recipe-details-modal",
			"Add recipe",
			"/recipes/modal",
			"#recipe-details-form",
		),
		components.Modal("recipe-details-modal", "Add recipe",
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
				html.Th(
					html.Class("col-5"),
					gomponents.Text("Name"),
				),
				html.Th(
					html.Class("col-5"),
					gomponents.Text("Description"),
				),
				html.Th(
					html.Class("col-2"),
					gomponents.Text("Actions"),
				),
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
		html.Td(gomponents.Text(recipe.Description)),
		html.Td(
			components.ModalButton(
				"recipe-details-modal",
				"Edit",
				fmt.Sprintf("/recipes/modal/%s", recipe.ID.String()),
				"#recipe-details-form",
			),
			html.Button(
				html.Type("button"),
				html.Class("btn btn-danger"),
				gomponents.Text("Delete"),
				htmx.Delete(fmt.Sprintf("/recipes/%s", recipe.ID)),
				htmx.Swap("none"),
			),
		),
	)
}
