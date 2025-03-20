package templates

import (
	"fmt"
	"strings"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/google/uuid"
)

func Recipes(accountID uuid.UUID) gomponents.Node {
	return BasePage("Recipes", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Recipes"),
			),
			html.Div(
				html.A(
					html.Class("btn btn-primary"),
					html.Role("button"),
					html.Data("bs-toggle", "collapse"),
					html.Href("#recipe-filters"),
					gomponents.Text("Filters"),
				),
			),
			html.Form(
				html.ID("recipe-filters"),
				html.Class("card card-body collapse"),
				FormInput(
					"recipe-name",
					"Name",
					nil,
					html.Input(
						html.Name("name"),
						html.Class("form-control"),
					),
				),
				html.Div(
					html.Class("dropdown"),
					html.A(
						html.Class("btn btn-secondary dropdown-toggle"),
						html.Type("button"),
						html.Role("button"),
						html.Data("bs-toggle", "dropdown"),
						gomponents.Text("Visibility"),
					),
					html.Ul(
						html.Class("text-center dropdown-menu"),
						html.Li(
							html.Class("dropdown-item"),
							FormCheck("", data.VisibilityPublic, false, html.Input(
								html.Class("form-check-input"),
								html.Type("checkbox"),
								html.Value(data.VisibilityPublic),
								html.Name("visibility"),
								html.Checked(),
							)),
							FormCheck("", data.VisibilityFriends, false, html.Input(
								html.Class("form-check-input"),
								html.Type("checkbox"),
								html.Value(data.VisibilityFriends),
								html.Name("visibility"),
								html.Checked(),
							)),
							FormCheck("", data.VisibilityPrivate, false, html.Input(
								html.Class("form-check-input"),
								html.Type("checkbox"),
								html.Value(data.VisibilityPrivate),
								html.Name("visibility"),
								html.Checked(),
							)),
						),
					),
				),
				htmx.Post("/recipes/search"),
				htmx.Target("#recipe-table"),
				htmx.Swap("innerHTML"),
				htmx.Vals(fmt.Sprintf(`{"accountID": "%s"}`, accountID)),
				htmx.Trigger("load,change delay:500ms,input changed delay:500ms,recipe-update from:body"),
			),
			html.Div(html.ID("recipe-table")),
			ModalButton(
				"btn-primary",
				"Add recipe",
				htmx.Get("/recipes//details"),
			),
		),
	})
}

func FavoriteRecipes() gomponents.Node {
	return BasePage("Favorites", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Favorites"),
			),
			html.Div(
				html.A(
					html.Class("btn btn-primary"),
					html.Role("button"),
					html.Data("bs-toggle", "collapse"),
					html.Href("#recipe-filters"),
					gomponents.Text("Filters"),
				),
			),
			html.Form(
				html.ID("recipe-filters"),
				html.Class("card card-body collapse"),
				FormInput(
					"recipe-name",
					"Name",
					nil,
					html.Input(
						html.Name("name"),
						html.Class("form-control"),
					),
				),
				htmx.Post("/recipes/search"),
				htmx.Target("#recipe-table"),
				htmx.Swap("innerHTML"),
				htmx.Vals(fmt.Sprintf(`{
					"favorites": true,
					"visibility": ["%s", "%s", "%s"]
				}`, data.VisibilityPublic, data.VisibilityFriends, data.VisibilityPrivate)),
				htmx.Trigger("load,change delay:500ms,input changed delay:500ms,recipe-update from:body"),
			),
			html.Div(html.ID("recipe-table")),
		),
	})
}

func ExploreRecipes() gomponents.Node {
	return BasePage("Explore", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Explore"),
			),
			html.Div(
				html.H3(gomponents.Text("Search recipes")),
				html.Input(
					html.Class("form-control"),
					html.Type("search"),
					html.Name("name"),
					html.Placeholder("Begin typing to seach recipes"),
					htmx.Post("/recipes/search"),
					htmx.Trigger("input changed delay:500ms, keyup[key=='Enter']"),
					htmx.Target("#recipes-search-table"),
					htmx.Vals(fmt.Sprintf(`{
						"visibility": ["%s", "%s", "%s"]
					}`, data.VisibilityPublic, data.VisibilityFriends, data.VisibilityPrivate)),
					htmx.Indicator(".htmx-indicator"),
				),
				html.Span(html.Class("htmx-indicator"), gomponents.Text("Searching...")),
				html.Div(html.ID("recipes-search-table")),
			),
		),
	})
}

func RecipeDetailsModalContent(accountID uuid.UUID, recipe data.Recipe, copy bool) gomponents.Node {
	viewOnly := recipe.ListID != uuid.Nil && recipe.AccountID != accountID && !copy

	return ModalContent(
		"Recipe details",
		gomponents.Group{
			gomponents.If(viewOnly, RecipeDetailsView(recipe)),
			gomponents.If(!viewOnly, RecipeDetailsEdit(recipe, copy)),
		},
		gomponents.Group{
			ModalDismiss(),
			gomponents.If(!viewOnly, ModalSubmit()),
		},
	)
}

func RecipeDetailsView(recipe data.Recipe) gomponents.Node {
	return html.Div(
		html.Class("text-center"),
		html.H2(gomponents.Text(recipe.Name)),
		html.P(gomponents.Text(recipe.Description)),
		gomponents.If(recipe.InstructionType != data.InstructionTypeNone,
			html.Div(
				gomponents.Iff(recipe.InstructionType == data.InstructionTypeText, func() gomponents.Node {
					instructionNodes := gomponents.Group{}
					for line := range strings.Lines(recipe.Instructions) {
						instructionNodes = append(instructionNodes, html.Span(gomponents.Text(line)), html.Br())
					}
					return html.Div(
						html.Class("card"),
						html.H4(gomponents.Text("Instructions")),
						html.Hr(),
						html.Div(
							html.Class("text-start"),
							instructionNodes,
						),
					)
				}),
				gomponents.If(recipe.InstructionType == data.InstructionTypeLink, html.A(
					html.Href(recipe.Instructions),
					html.Target("_blank"),
					html.Class("btn btn-primary"),
					html.Type("button"),
					gomponents.Text("View instructions on site"),
				)),
			),
		),
	)
}

func RecipeDetailsEdit(recipe data.Recipe, copy bool) gomponents.Node {
	ifExists := func(node gomponents.Node) gomponents.Node {
		return gomponents.If(recipe.ListID != uuid.Nil, node)
	}

	return ModalForm(
		gomponents.If(!copy, htmx.Post("/recipes")),
		gomponents.If(copy, htmx.Post(fmt.Sprintf("/recipes/%s/copy", recipe.ListID))),
		ifExists(html.Input(
			html.Type("hidden"),
			html.Name("id"),
			html.Value(recipe.ListID.String()),
		)),
		FormInput("recipe-name", "Recipe name", nil, html.Input(
			html.ID("recipe-name"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("name"),
			ifExists(html.Value(recipe.Name)),
			html.Required(),
		)),
		FormInput("recipe-description", "Recipe description", nil, html.Input(
			html.ID("recipe-description"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("description"),
			ifExists(html.Value(recipe.Description)),
		)),
		Select("recipeVisibility", "Visibility", "visibility", recipe.Visibility, []string{
			data.VisibilityPublic,
			data.VisibilityFriends,
			data.VisibilityPrivate,
		}, nil),
		html.Div(
			gomponents.Attr("x-data", fmt.Sprintf("{instructionType: '%s'}", recipe.InstructionType)),
			html.Div(
				html.Class("input-group"),
				html.Span(
					html.Class("input-group-text"),
					gomponents.Text("Recipe instructions"),
				),
				Select(
					"recipe-instruction-type",
					"Instruction type",
					"instruction-type",
					recipe.InstructionType,
					[]string{
						data.InstructionTypeNone,
						data.InstructionTypeText,
						data.InstructionTypeLink,
					},
					gomponents.Attr("x-on:change", "instructionType = $event.target.value"),
				),
			),
			html.Textarea(
				gomponents.Attr("x-show", "instructionType == 'text'"),
				gomponents.Attr("x-bind:disabled", "instructionType != 'text'"),
				html.Class("form-control"),
				html.Name("instructions"),
				html.Rows("10"),
				html.Required(),
				gomponents.If(recipe.InstructionType == data.InstructionTypeText, gomponents.Text(recipe.Instructions)),
			),
			html.Input(
				gomponents.Attr("x-show", "instructionType == 'link'"),
				gomponents.Attr("x-bind:disabled", "instructionType != 'link'"),
				html.Class("form-control"),
				html.Type("url"),
				html.Name("instructions"),
				gomponents.If(recipe.InstructionType == data.InstructionTypeLink, html.Value(recipe.Instructions)),
				html.Required(),
			),
		),
	)
}

func FormInput(id, label string, attributes, input gomponents.Node) gomponents.Node {
	return html.Div(
		html.Class("form-floating"),
		attributes,
		input,
		html.Label(
			html.For(id),
			gomponents.Text(label),
		),
	)
}

func FormCheck(id, label string, isSwitch bool, input gomponents.Node) gomponents.Node {
	class := "form-check"
	if isSwitch {
		class += " form-switch"
	}
	return html.Div(
		html.Class(class),
		input,
		html.Label(
			html.Class("form-check-label"),
			html.For(id),
			gomponents.Text(label),
		),
	)
}

func Disabled(b bool) gomponents.Node {
	if b {
		return html.Disabled()
	}
	return nil
}

func RecipeTable(accountID uuid.UUID, recipes []data.Recipe) gomponents.Node {
	var recipeRows gomponents.Group
	for _, recipe := range recipes {
		recipeRows = append(recipeRows, RecipeRow(accountID, recipe))
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

func RecipeRow(accountID uuid.UUID, recipe data.Recipe) gomponents.Node {
	actions := gomponents.Group{
		html.Li(
			html.Class("dropdown-item"),
			ModalButton(
				"btn-secondary w-100",
				"Details",
				htmx.Get(fmt.Sprintf("/recipes/%s/details", recipe.ListID.String())),
			),
		),
		html.Li(
			html.Class("dropdown-item"),
			html.A(
				html.Href(fmt.Sprintf("/lists/%v/ingredients", recipe.ListID)),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-secondary w-100"),
					gomponents.Text("Ingredients"),
				),
			),
		),
		FavoriteButton(recipe.ListID, recipe.Favorite),
		html.Li(
			html.Class("dropdown-item"),
			ModalButton(
				"btn-secondary w-100",
				"Copy",
				htmx.Get(fmt.Sprintf("/recipes/%s/copy", recipe.ListID.String())),
			),
		),
	}
	if accountID == recipe.AccountID {
		actions = append(actions,
			html.Li(html.Hr(html.Class("dropdown-divider"))),
			html.Li(
				html.Class("dropdown-item"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-danger w-100"),
					gomponents.Text("Delete"),
					htmx.Delete(fmt.Sprintf("/recipes/%v", recipe.ListID)),
					htmx.Swap("none"),
					htmx.Confirm("Are you sure you want to delete this recipe?"),
				),
			),
		)
	}

	return html.Tr(
		html.Td(gomponents.Text(recipe.Name)),
		html.Td(
			// Hide if the screen is small
			html.Class("d-none d-sm-table-cell"),
			gomponents.Text(recipe.Description)),
		html.Td(
			html.Div(
				html.Class("btn-group dropdown-center"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary"),
					gomponents.Text("Add to cart"),
					htmx.Post(fmt.Sprintf("/cart/list/%v", recipe.ListID)),
					htmx.Swap("none"),
				),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary dropdown-toggle dropdown-toggle-split"),
					html.Data("bs-toggle", "dropdown"),
				),
				html.Ul(
					html.Class("dropdown-menu"),
					actions,
				),
			),
		),
	)
}

func FavoriteButton(listID uuid.UUID, favorite bool) gomponents.Node {
	if favorite {
		return html.Li(
			html.Class("dropdown-item"),
			html.Button(
				html.Role("button"),
				html.Class("btn btn-secondary w-100"),
				gomponents.Text("Unfavorite"),
				htmx.Delete(fmt.Sprintf("/recipes/%v/favorite", listID)),
			),
		)
	}
	return html.Li(
		html.Class("dropdown-item"),
		html.Button(
			html.Role("button"),
			html.Class("btn btn-secondary w-100"),
			gomponents.Text("Favorite"),
			htmx.Post(fmt.Sprintf("/recipes/%v/favorite", listID)),
		),
	)
}
