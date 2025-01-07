package pages

import "maragu.dev/gomponents"

func DashboardPage() gomponents.Node {
	return BasePage("Dashboard", "/", gomponents.Group{})
}
