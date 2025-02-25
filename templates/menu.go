package templates

import (
	_ "embed"

	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

//go:embed theme.js
var ThemeJavascript string

func Menu() gomponents.Node {
	return gomponents.Group{
		html.Nav(
			html.Class("navbar bg-body-tertiary sticky-top"),
			html.Div(
				html.Class("container-fluid"),
				html.Button(
					html.Class("navbar-toggler"),
					html.Type("button"),
					html.Data("bs-toggle", "offcanvas"),
					html.Data("bs-target", "#offcanvasNavbar"),
					html.Aria("controls", "offcanvasNavbar"),
					html.Aria("label", "Toggle navigation"),
					html.Span(html.Class("navbar-toggler-icon")),
				),
				html.A(
					html.Class("navbar-brand"),
					html.Href("#"),
					gomponents.Text("Grocer Recipe Hub"),
				),
				html.Div(
					html.Class("offcanvas offcanvas-start"),
					html.TabIndex("-1"),
					html.ID("offcanvasNavbar"),
					html.Aria("labelledby", "offcanvasNavbarLabel"),
					html.Div(
						html.Class("offcanvas-header d-flex"),
						html.Div(
							html.Class("dropdown"),
							html.Button(
								html.Class("btn btn-link nav-link py-2 px-0 px-lg-2 dropdown-toggle d-flex align-items-center show"),
								html.ID("bd-theme"),
								html.Type("button"),
								html.Aria("expanded", "true"),
								html.Data("bs-toggle", "dropdown"),
								html.Data("bs-display", "static"),
								html.Aria("label", "Toggle theme (dark)"),
								html.I(html.Class("bi my-1 theme-icon-active")),
								html.Span(html.Class("d-none ms-2"), html.ID("bd-theme-text"), gomponents.Text("Toggle theme")),
							),
							html.Ul(
								html.Class("dropdown-menu"),
								html.Aria("labelledby", "bd-theme-text"),
								html.Data("bs-popper", "static"),
								html.Li(
									html.Button(
										html.Type("button"),
										html.Class("dropdown-item d-flex align-items-center"),
										html.Data("bs-theme-value", "light"),
										html.Aria("pressed", "false"),
										html.I(html.Class("bi-sun-fill me-2"), html.Data("icon", "bi-sun-fill")),
										gomponents.Text("Light"),
										html.I(html.Class("bi-check2 ms-auto d-none")),
									),
								),
								html.Li(
									html.Button(
										html.Type("button"),
										html.Class("dropdown-item d-flex align-items-center active"),
										html.Data("bs-theme-value", "dark"),
										html.Aria("pressed", "true"),
										html.I(html.Class("bi-moon-stars-fill me-2"), html.Data("icon", "bi-moon-stars-fill")),
										gomponents.Text("Dark"),
										html.I(html.Class("bi-check2 ms-auto d-none")),
									),
								),
								html.Li(
									html.Button(
										html.Type("button"),
										html.Class("dropdown-item d-flex align-items-center"),
										html.Data("bs-theme-value", "auto"),
										html.Aria("pressed", "false"),
										html.I(html.Class("bi-circle-half me-2"), html.Data("icon", "bi-circle-half")),
										gomponents.Text("Auto"),
										html.I(html.Class("bi-check2 ms-auto d-none")),
									),
								),
							),
						),
						html.H5(
							html.Class("offcanvas-title flex-fill text-center"),
							html.ID("offcanvasNavbarLabel"),
							gomponents.Text("Menu"),
						),
						html.Button(
							html.Type("button"),
							html.Class("btn-close"),
							html.Data("bs-dismiss", "offcanvas"),
							html.Aria("label", "Close"),
						),
					),
					html.Hr(html.Class("w-auto m-2")),
					html.Div(
						html.Class("offcanvas-body"),
						html.Div(
							html.Class("d-flex flex-column align-items-start h-100"),
							html.Div(
								html.Class("flex-grow-1 align-self-center w-75"),
								html.A(
									html.Class("btn btn-secondary w-100 my-2"),
									html.Href("/"),
									gomponents.Text("Home"),
								),
								html.Div(
									html.Class("dropdown"),
									html.Button(
										html.Type("button"),
										html.Class("btn btn-secondary dropdown-toggle w-100 my-2"),
										html.Data("bs-toggle", "dropdown"),
										gomponents.Text("Recipes"),
									),
									html.Ul(
										html.Class("dropdown-menu"),
										html.Li(html.A(html.Class("dropdown-item"), html.Href("/recipes"), gomponents.Text("My recipes"))),
										html.Li(html.A(html.Class("dropdown-item"), html.Href("/recipes/favorites"), gomponents.Text("Favorites"))),
										html.Li(html.A(html.Class("dropdown-item"), html.Href("/recipes/explore"), gomponents.Text("Explore"))),
									),
								),
								html.A(
									html.Class("btn btn-secondary w-100 my-2"),
									html.Href("/accounts/profiles"),
									gomponents.Text("Profiles"),
								),
								html.A(
									html.Class("btn btn-secondary w-100 my-2"),
									html.Href("/cart"),
									gomponents.Text("Cart"),
								),
							),
							html.Hr(html.Class("align-self-center w-75")),
							html.Div(
								html.Class("align-self-center w-75"),
								html.A(html.Class("btn btn-secondary w-100 my-2"), html.Href("/accounts"), gomponents.Text("Account")),
								// html.A(html.Class("btn btn-warning w-100 my-2"), html.Href("/users"), gomponents.Text("Admin")),
								html.Button(html.Class("btn btn-danger w-100 my-2"), htmx.Post("/auth/logout"), htmx.Swap("none"), gomponents.Text("Logout")),
							),
						),
					),
				),
			),
		),
		html.Script(gomponents.Raw(ThemeJavascript)),
	}
}
