package templates

import (
	"fmt"
	"math"

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
		Modal("cart-product-details", "Edit product", htmx.Put("/cart/product")),
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
			html.Step("0.01"),
			html.Required(),
			html.Value(fmt.Sprintf("%.2f", float64(cartProduct.Quantity)/100)),
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
	Staple      bool
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
	actions := gomponents.Group{
		ModalButton(
			"cart-product-details-modal",
			"Edit details",
			"",
			fmt.Sprintf("/cart/%v", cartProduct.ProductID),
			"#cart-product-details-form",
		),
	}
	if cartProduct.Staple {
		actions = append(actions,
			html.Button(
				html.Type("button"),
				html.Class("btn btn-primary"),
				gomponents.Text("Include"),
				htmx.Post(fmt.Sprintf("/cart/%v/include", cartProduct.ProductID)),
				htmx.Swap("none"),
			),
		)

	} else {
		actions = append(actions,
			html.Button(
				html.Type("button"),
				html.Class("btn btn-danger"),
				gomponents.Text("Delete"),
				htmx.Delete(fmt.Sprintf("/cart/%v", cartProduct.ProductID)),
				htmx.Swap("none"),
				htmx.Confirm("Are you sure you want to remove this product from your cart?"),
			),
		)

	}
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
		html.Td(gomponents.Text(fmt.Sprintf("%.2f -> %d", float64(cartProduct.Quantity)/100, int(math.Ceil(float64(cartProduct.Quantity)/100))))),
		html.Td(actions),
	)
}
