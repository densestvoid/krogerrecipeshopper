package templates

import (
	"fmt"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"github.com/google/uuid"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

type Account struct {
	ID        uuid.UUID
	ImageSize string
	Location  *data.CacheLocation
}

func AccountPage(account Account, profile *data.Profile) gomponents.Node {
	return BasePage("Account", "/", gomponents.Group{
		html.Div(
			html.Class("d-flex justify-content-center"),
			html.Div(
				html.Class("text-center"),
				html.H1(
					html.Class("m-2"),
					gomponents.Text("Account"),
				),
				html.Div(
					htmx.Get(fmt.Sprintf("/accounts/%s/profile", account.ID)),
					htmx.Trigger("profile-update from:body"),
					Profile(account, profile),
				),
				Settings(account),
				html.Button(
					html.Type("button"),
					html.Class("btn btn-danger"),
					gomponents.Text("Delete account"),
					htmx.Delete(fmt.Sprintf("/accounts/%s", account.ID)),
					htmx.Swap("none"),
					htmx.Confirm("Are you sure you want to delete this account? All data related to this account will be permanently deleted, and unrecoverable. If you sign back in with the same Kroger account, a new account will be created."),
				),
			),
		),
	})
}

func Profile(account Account, profile *data.Profile) gomponents.Node {
	if profile == nil {
		return html.Div(
			html.Class("card m-1"),
			html.Form(
				htmx.Post(fmt.Sprintf("/accounts/%s/profile", account.ID)),
				htmx.Swap("none"),
				html.Div(
					html.Class("card-header"),
					gomponents.Text("Create profile"),
				),
				html.Div(
					html.Class("card-body"),
					FormInput(
						"profile-display-name",
						"Display name",
						nil,
						html.Input(
							html.ID("profile-display-name"),
							html.Class("form-control"),
							html.Required(),
							html.Type("text"),
							html.MinLength("6"),
							html.Name("displayName"),
						),
					),
				),
				html.Div(
					html.Class("card-footer"),
					html.Button(
						html.Type("submit"),
						html.Class("btn btn-primary"),
						gomponents.Text("Create profile"),
					),
				),
			),
		)
	}
	return html.Div(
		html.Class("card m-1"),
		html.Form(
			htmx.Patch(fmt.Sprintf("/accounts/%s/profile", account.ID)),
			htmx.Swap("none"),
			html.Div(
				html.Class("card-header"),
				gomponents.Text("Update profile"),
			),
			html.Div(
				html.Class("card-body"),
				FormInput(
					"profile-display-name",
					"Display name",
					nil,
					html.Input(
						html.ID("profile-display-name"),
						html.Class("form-control"),
						html.Required(),
						html.Type("text"),
						html.MinLength("6"),
						html.Name("displayName"),
						html.Value(profile.DisplayName),
					),
				),
			),
			html.Div(
				html.Class("card-footer"),
				html.Button(
					html.Type("submit"),
					html.Class("btn btn-primary"),
					gomponents.Text("Update profile"),
				),
				html.Button(
					html.Class("btn btn-danger"),
					htmx.Delete(fmt.Sprintf("/accounts/%s/profile", account.ID)),
					htmx.Swap("none"),
					htmx.Confirm(`Are you sure you want to delete your profile?
This will remove all of your friends and your profile page where users can view all your recipes.
Your recipes will still be searchable on the explore page.`),
					gomponents.Text("Delete"),
				),
			),
		),
	)
}

func Settings(account Account) gomponents.Node {
	return html.Div(
		html.Class("card m-1"),
		html.Div(
			html.Class("card-header"),
			gomponents.Text("Update settings"),
		),
		html.Div(
			html.Class("card-body"),
			html.Div(
				html.Class("row"),
				html.Div(
					html.Class("col-3 align-content-center text-nowrap"),
					html.P(gomponents.Text("Location:")),
				),
				html.Div(
					html.ID("location-info"),
					html.Class("col-9 text-nowrap"),
					LocationNode(account.Location),
				),
			),
			ModalButton(
				"btn-secondary",
				"Change location",
				htmx.Get("/locations/details"),
			),
			html.Button(
				html.Class("btn btn-danger"),
				gomponents.Text("Clear"),
				htmx.Patch(fmt.Sprintf("/accounts/%s/settings", account.ID)),
				htmx.Target("#location-info"),
				htmx.Vals(`{"locationID": ""}`),
			),
			html.Form(
				html.ID("settings-form"),
				htmx.Patch(fmt.Sprintf("/accounts/%s/settings", account.ID)),
				htmx.Swap("none"),
				Select("accountImageSize", "Image size", "imageSize", account.ImageSize, []string{
					data.ImageSizeThumbnail,
					data.ImageSizeSmall,
					data.ImageSizeMedium,
					data.ImageSizeLarge,
					data.ImageSizeExtraLarge,
				}, nil),
			),
		),
		html.Div(
			html.Class("card-footer"),
			html.Button(
				html.Type("submit"),
				html.Class("btn btn-primary"),
				gomponents.Text("Save settings"),
				html.FormAttr("settings-form"),
			),
		),
	)
}

func LocationNode(location *data.CacheLocation) gomponents.Node {
	if location == nil {
		return html.P(gomponents.Text("None"))
	}
	return gomponents.Group{
		html.P(gomponents.Text(location.Name)),
		html.P(gomponents.Text(location.Address)),
	}
}

func Select[T comparable](id, label, name string, selected T, values []T, action gomponents.Node) gomponents.Node {
	var options gomponents.Group
	for _, value := range values {
		options = append(options, html.Option(
			gomponents.If(value == selected, html.Selected()),
			html.Value(fmt.Sprintf("%v", value)),
			gomponents.Textf("%v", value)))
	}
	return html.Div(
		html.Class("form-floating"),
		html.Select(
			html.ID(id),
			html.Class("form-select"),
			html.Name(name),
			action,
			options,
		),
		html.Label(
			html.For(id),
			gomponents.Text(label),
		),
		html.Required(),
	)
}

func Profiles(profiles []data.Profile) gomponents.Node {
	var profileButtons gomponents.Group
	for _, profile := range profiles {
		profileButtons = append(profileButtons, html.A(
			html.Class("flex-row align-self-center p-1"),
			html.Href(fmt.Sprintf("/profiles/%s", profile.AccountID)),
			html.Button(
				html.Class("btn btn-secondary w-100 text-nowrap"),
				gomponents.Text(profile.DisplayName),
			),
		))
	}

	return BasePage("Profiles", "/", gomponents.Group{
		html.H1(
			html.Class("text-center m-2"),
			gomponents.Text("Profiles"),
		),
		html.Div(
			html.Class("d-flex container justify-content-center"),
			profileButtons,
		),
	})
}

func ProfilePage(profile data.Profile) gomponents.Node {
	return BasePage(profile.DisplayName, "/", gomponents.Group{
		html.Div(
			html.Class("text-center"),
			html.H3(
				gomponents.Textf("%s's recipes", profile.DisplayName),
			),
			html.Div(
				htmx.Post("/recipes/search"),
				htmx.Swap("innerHTML"),
				htmx.Vals(fmt.Sprintf(`{
					"accountID": "%s",
					"visibility": ["%s", "%s"]
				}`, profile.AccountID, data.VisibilityPublic, data.VisibilityFriends)),
				htmx.Trigger("load"),
			),
		),
	})
}
