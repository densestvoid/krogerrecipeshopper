package templates

import (
	"fmt"

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
				html.Required(),
				Disabled(accountID != recipe.AccountID),
			)),
			FormInput("recipe-description", "Recipe description", html.Input(
				html.ID("recipe-description"),
				html.Class("form-control"),
				html.Type("text"),
				html.Name("description"),
				html.Value(recipe.Description),
				html.Required(),
				Disabled(accountID != recipe.AccountID),
			)),
			gomponents.If(accountID == recipe.AccountID, Select("recipeVisibility", "Visibility", "visibility", recipe.Visibility, []string{
				data.VisibilityPublic,
				data.VisibilityFriends,
				data.VisibilityPrivate,
			})),
		}
	}
	return gomponents.Group{
		FormInput("recipe-name", "Recipe name", html.Input(
			html.ID("recipe-name"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("name"),
			html.Required(),
			htmx.Validate("true"), // Recuired due to the use of hx-include
		)),
		FormInput("recipe-description", "Recipe description", html.Input(
			html.ID("recipe-description"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("description"),
			html.Required(),
			htmx.Validate("true"), // Recuired due to the use of hx-include
		)),
		Select("recipeVisibility", "Visibility", "visibility", data.VisibilityPublic, []string{
			data.VisibilityPrivate,
			data.VisibilityFriends,
			data.VisibilityPublic,
		}),
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

func FormCheck(id, label string, input gomponents.Node) gomponents.Node {
	return html.Div(
		html.Class("form-check"),
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
					htmx.Confirm("Are you sure you want to delete thi recipe?"),
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
