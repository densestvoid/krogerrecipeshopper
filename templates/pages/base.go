package pages

import (
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"

	"github.com/densestvoid/krogerrecipeshopper/templates/components"
)

func BasePage(title, baseURL string, bodyNodes gomponents.Group) gomponents.Node {
	return html.Doctype(
		html.HTML(
			html.Lang("en"),
			html.Data("bs-theme", "dark"),
			baseHead(title, baseURL),
			baseBody(bodyNodes),
		),
	)
}

func baseHead(title, baseURL string) gomponents.Node {
	return html.Head(
		html.Title(title),
		html.Meta(html.Charset("UTF-8")),
		html.Meta(html.Name("viewport")),
		// Bootstrap CSS
		html.Link(
			html.Href("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css"),
			html.Rel("stylesheet"),
			html.Integrity("sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH"),
			html.CrossOrigin("anonymous"),
		),
		// Bootstrap Icons
		html.Link(
			html.Href("https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css"),
			html.Rel("stylesheet"),
		),
		// Relative URLs base
		html.Base(html.Href(baseURL)),
	)
}

func baseBody(bodyNodes []gomponents.Node) gomponents.Node {
	bodyNodes = append(bodyNodes)
	return html.Body(
		// Menu
		components.Menu(),

		// Custom page content
		gomponents.Group(bodyNodes),

		// HTMX
		html.Script(html.Src("https://unpkg.com/htmx.org@2.0.4")),
		// Bootstrap JS
		html.Script(
			html.Src("https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js"),
			html.Integrity("sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz"),
			html.CrossOrigin("anonymous"),
		),
	)
}
