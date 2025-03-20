package templates

import (
	"fmt"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/google/uuid"
)

func Lists(accountID uuid.UUID) gomponents.Node {
	return BasePage("Lists", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Lists"),
			),
			html.Div(
				html.A(
					html.Class("btn btn-primary"),
					html.Role("button"),
					html.Data("bs-toggle", "collapse"),
					html.Href("#list-filters"),
					gomponents.Text("Filters"),
				),
			),
			html.Form(
				html.ID("list-filters"),
				html.Class("card card-body collapse"),
				FormInput(
					"list-name",
					"Name",
					nil,
					html.Input(
						html.Name("name"),
						html.Class("form-control"),
					),
				),
				htmx.Post("/lists/search"),
				htmx.Target("#list-table"),
				htmx.Swap("innerHTML"),
				htmx.Vals(fmt.Sprintf(`{"accountID": "%s"}`, accountID)),
				htmx.Trigger("load,change delay:500ms,input changed delay:500ms,list-update from:body"),
			),
			html.Div(html.ID("list-table")),
			ModalButton(
				"btn-primary",
				"Add list",
				htmx.Get("/lists//details"),
			),
		),
	})
}

func ListDetailsModalContent(accountID uuid.UUID, list data.List, copy bool) gomponents.Node {
	viewOnly := list.ID != uuid.Nil && list.AccountID != accountID && !copy

	return ModalContent(
		"List details",
		gomponents.Group{
			gomponents.If(viewOnly, ListDetailsView(list)),
			gomponents.If(!viewOnly, ListDetailsEdit(list, copy)),
		},
		gomponents.Group{
			ModalDismiss(),
			gomponents.If(!viewOnly, ModalSubmit()),
		},
	)
}

func ListDetailsView(list data.List) gomponents.Node {
	return html.Div(
		html.Class("text-center"),
		html.H2(gomponents.Text(list.Name)),
		html.P(gomponents.Text(list.Description)),
	)
}

func ListDetailsEdit(list data.List, copy bool) gomponents.Node {
	ifExists := func(node gomponents.Node) gomponents.Node {
		return gomponents.If(list.ID != uuid.Nil, node)
	}

	return ModalForm(
		gomponents.If(!copy, htmx.Post("/lists")),
		gomponents.If(copy, htmx.Post(fmt.Sprintf("/lists/%s/copy", list.ID))),
		ifExists(html.Input(
			html.Type("hidden"),
			html.Name("id"),
			html.Value(list.ID.String()),
		)),
		FormInput("list-name", "List name", nil, html.Input(
			html.ID("list-name"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("name"),
			ifExists(html.Value(list.Name)),
			html.Required(),
		)),
		FormInput("list-description", "List description", nil, html.Input(
			html.ID("list-description"),
			html.Class("form-control"),
			html.Type("text"),
			html.Name("description"),
			ifExists(html.Value(list.Description)),
		)),
	)
}

func ListTable(accountID uuid.UUID, lists []data.List) gomponents.Node {
	var listRows gomponents.Group
	for _, list := range lists {
		listRows = append(listRows, ListRow(accountID, list))
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
			listRows,
		),
	)
}

func ListRow(accountID uuid.UUID, list data.List) gomponents.Node {
	actions := gomponents.Group{
		html.Li(
			html.Class("dropdown-item"),
			ModalButton(
				"btn-secondary w-100",
				"Details",
				htmx.Get(fmt.Sprintf("/lists/%s/details", list.ID.String())),
			),
		),
		html.Li(
			html.Class("dropdown-item"),
			html.A(
				html.Href(fmt.Sprintf("/lists/%v/ingredients", list.ID)),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-secondary w-100"),
					gomponents.Text("Ingredients"),
				),
			),
		),
		html.Li(
			html.Class("dropdown-item"),
			ModalButton(
				"btn-secondary w-100",
				"Copy",
				htmx.Get(fmt.Sprintf("/lists/%s/copy", list.ID.String())),
			),
		),
	}
	if accountID == list.AccountID {
		actions = append(actions,
			html.Li(html.Hr(html.Class("dropdown-divider"))),
			html.Li(
				html.Class("dropdown-item"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-danger w-100"),
					gomponents.Text("Delete"),
					htmx.Delete(fmt.Sprintf("/lists/%v", list.ID)),
					htmx.Swap("none"),
					htmx.Confirm("Are you sure you want to delete this list?"),
				),
			),
		)
	}

	return html.Tr(
		html.Td(gomponents.Text(list.Name)),
		html.Td(
			// Hide if the screen is small
			html.Class("d-none d-sm-table-cell"),
			gomponents.Text(list.Description)),
		html.Td(
			html.Div(
				html.Class("btn-group dropdown-center"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary"),
					gomponents.Text("Add to cart"),
					htmx.Post(fmt.Sprintf("/cart/list/%v", list.ID)),
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
