package templates

import (
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func ErrorPage(statusCode int, statusCodeText string) gomponents.Node {
	return BasePage("Error", "/", gomponents.Group{
		gomponents.If(400 <= statusCode && statusCode <= 499, html.Div(
			html.Class("d-flex flex-column min-vh-100 justify-content-center align-items-center"),
			html.H1(gomponents.Textf("%d - Client Error: %s", statusCode, statusCodeText)),
			html.P(gomponents.Text("Oops! It seems like there was a problem with the webpage.")),
			html.P(gomponents.Text("Please try again.")),
		)),
		gomponents.If(500 <= statusCode && statusCode <= 599, html.Div(
			html.Class("d-flex flex-column min-vh-100 justify-content-center align-items-center"),
			html.H1(gomponents.Textf("%d - Server Error: %s", statusCode, statusCodeText)),
			html.P(gomponents.Text("Oops! It seems like there was a problem with the server.")),
			html.P(gomponents.Text("Please try again later.")),
		)),
	})
}
