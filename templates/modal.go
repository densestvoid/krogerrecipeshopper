package templates

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func Modal() gomponents.Node {
	return html.Div(
		html.ID("modal"),
		html.Class("modal"),
		html.Div(
			html.Class("modal-dialog modal-lg"),
			html.Div(
				html.ID("modal-content"),
				html.Class("modal-content"),
			),
		),
	)
}

func ModalButton(btnClass, text string, htmxContentEndpoint gomponents.Node) gomponents.Node {
	return html.Button(
		html.Type("button"),
		html.Class(fmt.Sprintf("btn %s", btnClass)),
		html.Data("bs-toggle", "modal"),
		html.Data("bs-target", "#modal"),
		htmxContentEndpoint,
		htmx.Target("#modal-content"),
		htmx.Trigger("click"),
		gomponents.Text(text),
	)
}

func ModalContent(
	title string,
	bodyContent gomponents.Node,
	footerContent gomponents.Node,
) gomponents.Node {
	return gomponents.Group{
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
			html.Class("modal-body"),
			bodyContent,
		),
		html.Div(
			html.Class("modal-footer"),
			footerContent,
		),
	}
}

func ModalForm(htmxFormSubmit gomponents.Node, inputs ...gomponents.Node) gomponents.Node {
	return html.Form(
		html.ID("modal-form"),
		htmxFormSubmit,
		htmx.Swap("none"),
		// Required for bootrap and htmx to play nicely when submitting a form and closing a modal
		gomponents.Attr("hx-on::after-request", `if (event.target === this) {
			bootstrap.Modal.getInstance(this.closest('.modal')).hide();
		}`),
		gomponents.Group(inputs),
	)
}

func ModalDismiss() gomponents.Node {
	return html.Button(
		html.Class("btn btn-secondary"),
		html.Data("bs-dismiss", "modal"),
		html.Type("button"),
		html.Role("button"),
		gomponents.Text("Close"),
	)
}

func ModalSubmit() gomponents.Node {
	return html.Button(
		html.Class("btn btn-primary"),
		html.FormAttr("modal-form"),
		html.Role("button"),
		html.Type("submit"),
		gomponents.Text("Submit"),
	)
}
