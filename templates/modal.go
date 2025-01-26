package templates

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func ModalButton(id, text, classes, contentURL, contentTarget string) gomponents.Node {
	return html.Button(
		html.Type("button"),
		html.Class(fmt.Sprintf("btn btn-secondary %s", classes)),
		html.Data("bs-toggle", "modal"),
		html.Data("bs-target", fmt.Sprintf("#%s", id)),
		gomponents.Text(text),
		htmx.Get(contentURL),
		htmx.Target(contentTarget),
		htmx.Trigger("click"),
	)
}

func Modal(
	id string,
	title string,
	htmxFormSubmit gomponents.Node,
) gomponents.Node {
	return html.Div(
		html.Class("modal"),
		html.ID(fmt.Sprintf("%s-modal", id)),
		html.Div(
			html.Class("modal-dialog modal-lg"),
			html.Div(
				html.Class("modal-content"),
				html.Form(
					htmxFormSubmit,
					htmx.Swap("none"),
					// Required for bootrap and htmx to play nicely when submitting a form and closing a modal
					gomponents.Attr("hx-on::after-request", `if (event.target === this) {
						bootstrap.Modal.getInstance(this.closest('.modal')).hide();
					}`),
					html.Div(
						html.Class("modal-header"),
						html.H5(
							html.Class("modal-title"),
							gomponents.Text(title),
						),
						html.Button(
							html.Type("button"),
							html.Class("btn-close"),
							html.Data("bs-dismiss", "modal"),
							html.Aria("label", "Close"),
						),
					),
					html.Div(
						html.ID(fmt.Sprintf("%s-form", id)),
						html.Class("modal-body"),
					),
					html.Div(
						html.Class("modal-footer"),
						html.Button(
							html.Type("button"),
							html.Data("bs-dismiss", "modal"),
							html.Class("btn btn-secondary"),
							gomponents.Text("Close"),
						),
						gomponents.If(htmxFormSubmit != nil, html.Button(
							html.Role("button"),
							html.Type("submit"),
							html.Class("btn btn-primary"),
							gomponents.Text("Submit"),
						)),
					),
				),
			),
		),
	)
}
