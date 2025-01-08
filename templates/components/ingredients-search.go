package components

import (
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

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

func IngredientsTable() gomponents.Node {

}
