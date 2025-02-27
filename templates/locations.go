package templates

import (
	"fmt"

	"github.com/google/uuid"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func LocationsSearchModal(accountID uuid.UUID) gomponents.Node {
	return ModalContent(
		"Search locations",
		ModalForm(
			gomponents.Group{
				htmx.Patch(fmt.Sprintf("/accounts/%s/settings", accountID)),
				htmx.Swap("innerHTML"),
				htmx.Target("#location-info"),
			},

			// Locations search
			html.Div(
				html.Form(
					html.Input(
						html.Class("form-control"),
						html.Type("number"),
						html.Name("zipCode"),
						html.Placeholder("Enter the zip code of your local store"),
						htmx.Get("/locations"),
						htmx.Swap("innerHTML"),
						htmx.Trigger("input changed delay:1s, keyup[key=='Enter']"),
						htmx.Target("#locations-search-table"),
						htmx.Indicator(".htmx-indicator"),
					),
				),
			),

			// Loctions select
			html.Div(
				html.Span(html.Class("htmx-indicator"), gomponents.Text("Searching...")),
				html.Input(
					html.Class("d-none"),
					html.Type("radio"),
					html.Name("locationID"),
					html.Required(),
				),
				html.Div(html.ID("locations-search-table")),
			),
		),
		gomponents.Group{
			ModalDismiss(),
			ModalSubmit(),
		},
	)
}

type Location struct {
	LocationID string
	Name       string
	Address    string
}

func LocationsSearchTable(locations []Location) gomponents.Node {
	var locationRows gomponents.Group
	for _, location := range locations {
		locationRows = append(locationRows, LocationSearchRow(location))
	}
	return html.Table(
		html.THead(
			html.Tr(
				html.Th(gomponents.Text("Select")),
				html.Th(gomponents.Text("Name")),
				html.Th(gomponents.Text("Address")),
			),
		),
		html.TBody(locationRows),
	)
}

func LocationSearchRow(location Location) gomponents.Node {
	return html.Tr(
		html.Class("text-start"),
		html.Td(
			html.Input(
				html.Type("radio"),
				html.Name("locationID"),
				html.Value(location.LocationID),
			),
		),
		html.Td(gomponents.Text(location.Name)),
		html.Td(gomponents.Text(location.Address)),
	)
}
