package templates

import "maragu.dev/gomponents"

func DashboardPage() gomponents.Node {
	return BasePage("Dashboard", "/", gomponents.Group{})
}
