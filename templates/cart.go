package templates

import (
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func Cart() gomponents.Node {
	return BasePage("Cart", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Cart"),
			),
			html.Div(
				htmx.Get("/cart/table"),
				htmx.Swap("innerHTML"),
				htmx.Trigger("load,cart-update from:body"),
			),
			html.Div(
				html.Class("btn-group"),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary"),
					gomponents.Text("Send to "),
					html.Img(html.Src("https://developer.kroger.com/assets/logos/kroger.svg")),
					gomponents.Text(" cart"),
					htmx.Post("/cart/checkout"),
					htmx.Swap("none"),
					htmx.Trigger("click"),
				),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary dropdown-toggle dropdown-toggle-split"),
					html.Data("bs-toggle", "dropdown"),
				),
				html.Ul(
					html.Class("dropdown-menu"),
					html.Li(
						html.Class("dropdown-item"),
						html.A(
							html.Class("btn btn-primary"),
							html.Role("button"),
							html.Href("/shopping-list"),
							gomponents.Text("Shop in store"),
							gomponents.Attr("onclick", "return confirm('All non-included staple products will be removed. Continue?')"),
						),
					),
				),
			),
		),
	})
}

func CartProductDetailsModalContent(cartProduct data.CartProduct) gomponents.Node {
	return ModalContent(
		"Product Details",
		ModalForm(
			htmx.Put("/cart/product"),
			html.Input(
				html.Type("hidden"),
				html.Name("productID"),
				html.Value(cartProduct.ProductID),
			),
			FormInput("product-quantity", "product-quantity", nil, html.Input(
				html.ID("product-quantity"),
				html.Class("form-control"),
				html.Type("number"),
				html.Name("quantity"),
				html.Min("0.01"),
				html.Step("0.01"),
				html.Required(),
				html.Value(fmt.Sprintf("%.2f", float64(cartProduct.Quantity)/100)),
			)),
		),
		gomponents.Group{
			ModalDismiss(),
			ModalSubmit(),
		},
	)
}

type CartProduct struct {
	ProductID   string
	Brand       string
	Description string
	Size        string
	ImageURL    string
	Quantity    int
	Staple      bool
	ProductURL  string
	Location    string
}

func CartTable(cartProducts []CartProduct) gomponents.Node {
	var ingredientRows, stapleRows gomponents.Group
	for _, cartProduct := range cartProducts {
		if cartProduct.Staple {
			stapleRows = append(stapleRows, CartProductRow(cartProduct))
		} else {
			ingredientRows = append(ingredientRows, CartProductRow(cartProduct))
		}
	}
	return gomponents.Group{
		html.Table(
			html.Class("table table-striped table-bordered text-center align-middle w-100"),
			html.THead(
				html.Tr(
					html.Th(gomponents.Text("Product")),
					html.Th(gomponents.Text("Quantity")),
					html.Th(gomponents.Text("Actions")),
				),
			),
			html.TBody(
				html.Class("table-group-divider"),
				html.Tr(html.Td(html.ColSpan("6"), gomponents.Text("To order"))),
				ingredientRows,
			),
			html.TBody(
				html.Class("table-group-divider"),
				html.Tr(html.Td(html.ColSpan("6"), gomponents.Text("Add staples"))),
				stapleRows,
			),
		),
	}
}

func CartProductRow(cartProduct CartProduct) gomponents.Node {
	var primaryButton gomponents.Node
	if !cartProduct.Staple {
		primaryButton = ModalButton(
			"btn-primary",
			"Edit",
			htmx.Get(fmt.Sprintf("/cart/%v", cartProduct.ProductID)),
		)
	} else {
		primaryButton = html.Button(
			html.Type("button"),
			html.Class("btn btn-primary"),
			gomponents.Text("Include"),
			htmx.Post(fmt.Sprintf("/cart/%v/include", cartProduct.ProductID)),
			htmx.Swap("none"),
		)
	}

	return html.Tr(
		html.Td(
			html.Div(
				html.Class("d-flex flex-column align-items-center"),
				html.Img(
					html.Class("row img-fluid img-thumbnail"),
					html.Src(cartProduct.ImageURL),
				),
				html.Span(gomponents.Text(cartProduct.Brand)),
				html.A(
					html.Href(cartProduct.ProductURL),
					html.Target("_blank"),
					gomponents.Text(cartProduct.Description),
				),
				html.Span(gomponents.Text(cartProduct.Size)),
			),
		),
		html.Td(
			html.Div(
				html.Class("d-flex flex-column align-items-center"),
				html.Span(gomponents.Textf("%.2f", float64(cartProduct.Quantity)/100)),
				html.I(html.Class("bi bi-arrow-down")),
				html.Span(gomponents.Textf("%d", int(math.Ceil(float64(cartProduct.Quantity)/100)))),
			),
		),
		html.Td(
			html.Div(
				html.Class("btn-group dropdown-center"),
				primaryButton,
				html.Button(
					html.Type("button"),
					html.Class("btn btn-primary dropdown-toggle dropdown-toggle-split"),
					html.Data("bs-toggle", "dropdown"),
				),
				html.Ul(
					html.Class("dropdown-menu"),
					html.Li(
						html.Class("dropdown-item"),
						html.Button(
							html.Type("button"),
							html.Class("btn btn-danger w-100"),
							gomponents.Text("Delete"),
							htmx.Delete(fmt.Sprintf("/cart/%v", cartProduct.ProductID)),
							htmx.Swap("none"),
							htmx.Confirm("Are you sure you want to remove this product from your cart?"),
						),
					),
				),
			),
		),
	)
}

func ShoppingList() gomponents.Node {
	return BasePage("Shopping List", "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Text("Shopping List"),
			),
			html.Div(
				htmx.Get("/shopping-list/table"),
				htmx.Swap("innerHTML"),
				htmx.Trigger("load,cart-update from:body"),
			),
		),
	})
}

func ShoppingListTable(cartProducts []CartProduct) gomponents.Node {
	var locations = map[string]gomponents.Group{}
	for _, cartProduct := range cartProducts {
		locationGroup, ok := locations[cartProduct.Location]
		if !ok {
			locationGroup = gomponents.Group{}
			locations[cartProduct.Location] = locationGroup
		}
		locations[cartProduct.Location] = append(locations[cartProduct.Location], ShoppingListRow(cartProduct))
	}

	locationKeys := slices.Collect(maps.Keys(locations))
	slices.Sort(locationKeys)

	var locationGroups gomponents.Group
	for _, location := range locationKeys {
		locationGroups = append(locationGroups, ShoppingListLocation(location, locations[location]))
	}

	return gomponents.Group{
		html.Table(
			html.Class("table table-striped table-bordered text-center align-middle w-100"),
			html.THead(
				html.Tr(
					html.Th(gomponents.Text("Product")),
					html.Th(gomponents.Text("Quantity")),
					html.Th(gomponents.Text("Actions")),
				),
			),
			locationGroups,
		),
	}
}

func ShoppingListLocation(location string, shoppingListRows gomponents.Group) gomponents.Node {
	return html.TBody(
		html.Class("table-group-divider"),
		html.Tr(html.Td(html.ColSpan("6"), gomponents.Text(location))),
		shoppingListRows,
	)
}

func ShoppingListRow(cartProduct CartProduct) gomponents.Node {
	return html.Tr(
		html.Td(
			html.Div(
				html.Class("d-flex flex-column align-items-center"),
				html.Img(
					html.Class("row img-fluid img-thumbnail"),
					html.Src(cartProduct.ImageURL),
				),
				html.Span(gomponents.Text(cartProduct.Brand)),
				html.A(
					html.Href(cartProduct.ProductURL),
					html.Target("_blank"),
					gomponents.Text(cartProduct.Description),
				),
				html.Span(gomponents.Text(cartProduct.Size)),
			),
		),
		html.Td(
			html.Div(
				html.Class("d-flex flex-column align-items-center"),
				html.Span(gomponents.Textf("%.2f", float64(cartProduct.Quantity)/100)),
				html.I(html.Class("bi bi-arrow-down")),
				html.Span(gomponents.Textf("%d", int(math.Ceil(float64(cartProduct.Quantity)/100)))),
			),
		),
		html.Td(
			html.Button(
				html.Type("button"),
				html.Class("btn btn-primary w-100"),
				gomponents.Text("Check"),
				htmx.Delete(fmt.Sprintf("/cart/%v", cartProduct.ProductID)),
				htmx.Swap("none"),
			),
		),
	)
}
