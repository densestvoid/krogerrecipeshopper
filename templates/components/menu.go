package components

import (
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func Menu() gomponents.Node {
	return html.Nav(
		html.Ul(
			navItem("/", "Dashboard"),
			navItem("/explore", "Explore"),
			navItem("/recipes", "Recipes"),
			navItem("/favorites", "Favorites"),
			navItem("/cart", "Cart"),
			navItem("/account", "Account"),
		),
	)
}

func navItem(link, text string) gomponents.Node {
	return html.Li(html.A(html.Href(link), gomponents.Text(text)))
}
