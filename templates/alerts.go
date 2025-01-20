package templates

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func Alert(text, class string) gomponents.Node {
	return html.Div(
		html.ID("alerts"),
		htmx.SwapOOB("beforeend"),
		html.Div(
			html.Role("alert"),
			html.Class(fmt.Sprintf("alert %s fade show d-inline-flex m-auto", class)),
			gomponents.Attr("remove-me", "3s"),
			gomponents.Text(text),
			html.Button(
				html.Type("button"),
				html.Class("btn-close"),
				html.Data("bs-dismiss", "alert"),
			),
		),
	)
}
