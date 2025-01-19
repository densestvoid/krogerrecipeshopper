package templates

import (
	"fmt"

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
			html.Button(
				html.Type("button"),
				html.Class("btn btn-primary"),
				gomponents.Text("Submit"),
				htmx.Post("/cart/checkout"),
				htmx.Swap("none"),
				htmx.Trigger("click"),
			),
		),
		Modal("cart-product-details-modal", "Edit product",
			gomponents.Group{},
			gomponents.Group{
				html.Form(
					html.ID("cart-product-details-form"),
				),
			},
			gomponents.Group{
				html.Button(
					html.Type("submit"),
					html.Class("btn btn-primary"),
					html.Data("bs-dismiss", "modal"),
					gomponents.Text("Submit"),
					htmx.Include("#cart-product-details-form"),
					htmx.Put(fmt.Sprintf("/cart/product")),
					htmx.Swap("none"),
				),
			},
		),
	})
}

func CartProductDetailsForm(cartProduct data.CartProduct) gomponents.Node {
	return gomponents.Group{
		html.Input(
			html.Type("hidden"),
			html.Name("productID"),
			html.Value(cartProduct.ProductID),
		),
		FormInput("product-quantity", "product-quantity", html.Input(
			html.ID("product-quantity"),
			html.Class("form-control"),
			html.Type("number"),
			html.Name("quantity"),
			html.Min("0.01"),
			html.Step("0.5"),
			html.Value(fmt.Sprintf("%.2f", float32(cartProduct.Quantity)/100)),
		)),
	}
}

type CartProduct struct {
	ProductID   string
	Brand       string
	Description string
	Size        string
	ImageURL    string
	Quantity    int
}

func CartTable(cartProducts []CartProduct) gomponents.Node {
	var ingredientRows gomponents.Group
	for _, cartProduct := range cartProducts {
		ingredientRows = append(ingredientRows, CartProductRow(cartProduct))
	}
	return html.Table(
		html.Class("table table-striped table-bordered text-center align-middle w-100"),
		html.THead(
			html.Tr(
				html.Th(gomponents.Text("Image")),
				html.Th(gomponents.Text("Brand")),
				html.Th(gomponents.Text("Description")),
				html.Th(gomponents.Text("Size")),
				html.Th(gomponents.Text("Quantity")),
				html.Th(gomponents.Text("Actions")),
			),
		),
		html.TBody(
			html.Class("table-group-divider"),
			ingredientRows,
		),
	)
}

func CartProductRow(cartProduct CartProduct) gomponents.Node {
	return html.Tr(
		html.Td(
			html.Img(
				html.Class("img-fluid img-thumbnail"),
				html.Src(cartProduct.ImageURL),
			),
		),
		html.Td(gomponents.Text(cartProduct.Brand)),
		html.Td(gomponents.Text(cartProduct.Description)),
		html.Td(gomponents.Text(cartProduct.Size)),
		html.Td(gomponents.Text(fmt.Sprintf("%.2f", float32(cartProduct.Quantity)/100))),
		html.Td(
			ModalButton(
				"cart-product-details-modal",
				"Edit details",
				"",
				fmt.Sprintf("/cart/%v", cartProduct.ProductID),
				"#cart-product-details-form",
			),
			html.Button(
				html.Type("button"),
				html.Class("btn btn-danger"),
				gomponents.Text("Delete"),
				htmx.Delete(fmt.Sprintf("/cart/%v", cartProduct.ProductID)),
				htmx.Swap("none"),
			),
		),
	)
}
