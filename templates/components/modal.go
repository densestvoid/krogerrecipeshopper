package components

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func ModalButton(id, text, contentURL, contentTarget string) gomponents.Node {
	return html.Button(
		html.Type("button"),
		html.Class("btn btn-primary"),
		html.Data("bs-toggle", "modal"),
		html.Data("bs-target", fmt.Sprintf("#%s", id)),
		gomponents.Text(text),
		htmx.Get(contentURL),
		htmx.Target(contentTarget),
	)
}

func Modal(
	id string,
	title string,
	header gomponents.Group,
	body gomponents.Group,
	footer gomponents.Group,
) gomponents.Node {
	return html.Div(
		html.Class("modal"),
		html.ID(id),
		html.Div(
			html.Class("modal-dialog"),
			html.Div(
				html.Class("modal-content"),
				html.Div(
					html.Class("modal-header"),
					html.H5(
						html.Class("modal-title"),
						gomponents.Text(title),
					),
					header,
					html.Button(
						html.Type("button"),
						html.Class("btn-close"),
						html.Data("bs-dismiss", "modal"),
						html.Aria("label", "Close"),
					),
				),
				html.Div(
					html.Class("modal-body"),
					body,
				),
				html.Div(
					html.Class("modal-footer"),
					html.Button(
						html.Type("button"),
						html.Data("bs-dismiss", "modal"),
						html.Class("btn btn-secondary"),
						gomponents.Text("Close"),
					),
					footer,
				),
			),
		),
	)
}
