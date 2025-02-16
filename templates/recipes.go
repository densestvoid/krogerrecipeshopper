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
				htmx.Post("/recipes/search"),
				htmx.Swap("innerHTML"),
				htmx.Vals(fmt.Sprintf(`{"accountID": "%s"}`, accountID)),
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
		Modal("recipe-details", "Edit recipe", htmx.Post("/recipes")),
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
				htmx.Post("/recipes/search"),
				htmx.Swap("innerHTML"),
				htmx.Vals(`{"favorites": true}`),
				htmx.Trigger("load,recipe-update from:body"),
			),
		),
		Modal("recipe-details", "View recipe", nil),
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
					htmx.Indicator(".htmx-indicator"),
				),
				html.Span(html.Class("htmx-indicator"), gomponents.Text("Searching...")),
				html.Div(html.ID("recipes-search-table")),
			),
		),
		Modal("recipe-details", "View recipe", nil),
	})
}

func RecipeDetailsForm(accountID uuid.UUID, recipe *data.Recipe) gomponents.Node {
	if recipe != nil {
		if recipe.AccountID != accountID {
			return RecipeDetailsView(recipe)
		}
		return gomponents.Group{
			html.Input(
				html.Type("hidden"),
				html.Name("id"),
				html.Value(recipe.ID.String()),
			),
			FormInput("recipe-name", "Recipe name", nil, html.Input(
				html.ID("recipe-name"),
				html.Class("form-control"),
				html.Type("text"),
				html.Name("name"),
				html.Value(recipe.Name),
				html.Required(),
				Disabled(accountID != recipe.AccountID),
			)),
			FormInput("recipe-description", "Recipe description", nil, html.Input(
				html.ID("recipe-description"),
				html.Class("form-control"),
				html.Type("text"),
				html.Name("description"),
				html.Value(recipe.Description),
				Disabled(accountID != recipe.AccountID),
			)),
			gomponents.If(accountID == recipe.AccountID, Select("recipeVisibility", "Visibility", "visibility", recipe.Visibility, []string{
				data.VisibilityPublic,
				data.VisibilityFriends,
				data.VisibilityPrivate,
			}, nil)),
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
					Disabled(accountID != recipe.AccountID),
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
					Disabled(accountID != recipe.AccountID),
				),
			),
		}
	}
	return gomponents.Group{
		FormInput("recipe-name", "Recipe name", nil, html.Input(
			html.ID("recipe-name"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("name"),
			html.Required(),
		)),
		FormInput("recipe-description", "Recipe description", nil, html.Input(
			html.ID("recipe-description"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("description"),
		)),
		Select("recipeVisibility", "Visibility", "visibility", data.VisibilityPublic, []string{
			data.VisibilityPrivate,
			data.VisibilityFriends,
			data.VisibilityPublic,
		}, nil),
		html.Div(
			gomponents.Attr("x-data", "{instructionType : 'none'}"),
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
					data.InstructionTypeNone,
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
			),
			html.Input(
				gomponents.Attr("x-show", "instructionType == 'link'"),
				gomponents.Attr("x-bind:disabled", "instructionType != 'link'"),
				html.Class("form-control"),
				html.Type("url"),
				html.Name("instructions"),
				html.Required(),
			),
		),
	}
}

func RecipeDetailsView(recipe *data.Recipe) gomponents.Node {
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
			ModalButton(
				"recipe-details-modal",
				"Details",
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
					gomponents.Text("Ingredients"),
				),
			),
		),
		FavoriteButton(recipe.ID, recipe.Favorite),
	}
	if accountID == recipe.AccountID {
		actions = append(actions,
			html.Li(html.Hr(html.Class("dropdown-divider"))),
			html.Li(
				html.Button(
					html.Type("button"),
					html.Class("btn btn-danger w-100"),
					gomponents.Text("Delete"),
					htmx.Delete(fmt.Sprintf("/recipes/%v", recipe.ID)),
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
					actions,
				),
			),
		),
	)
}

func FavoriteButton(recipeID uuid.UUID, favorite bool) gomponents.Node {
	if favorite {
		return html.Li(
			html.Button(
				html.Role("button"),
				html.Class("btn btn-secondary w-100"),
				gomponents.Text("Unfavorite"),
				htmx.Delete(fmt.Sprintf("/recipes/%v/favorite", recipeID)),
			),
		)
	}
	return html.Li(
		html.Button(
			html.Role("button"),
			html.Class("btn btn-secondary w-100"),
			gomponents.Text("Favorite"),
			htmx.Post(fmt.Sprintf("/recipes/%v/favorite", recipeID)),
		),
	)
}
