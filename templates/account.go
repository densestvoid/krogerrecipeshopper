package templates

import (
	"fmt"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func Account(account data.Account, profile *data.Profile) gomponents.Node {
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

func Profile(account data.Account, profile *data.Profile) gomponents.Node {
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

func Settings(account data.Account) gomponents.Node {
	return html.Div(
		html.Class("card m-1"),
		html.Form(
			htmx.Patch(fmt.Sprintf("/accounts/%s/settings", account.ID)),
			htmx.Swap("none"),
			html.Div(
				html.Class("card-header"),
				gomponents.Text("Update settings"),
			),
			html.Div(
				html.Class("card-body"),
				Select("imageSize", account.ImageSize, []string{
					data.ImageSizeThumbnail,
					data.ImageSizeSmall,
					data.ImageSizeMedium,
					data.ImageSizeLarge,
					data.ImageSizeExtraLarge,
				}),
			),
			html.Div(
				html.Class("card-footer"),
				html.Button(
					html.Type("submit"),
					html.Class("btn btn-primary"),
					gomponents.Text("Save settings"),
				),
			),
		),
	)
}

func Select[T comparable](name string, selected T, values []T) gomponents.Node {
	var options gomponents.Group
	for _, value := range values {
		options = append(options, html.Option(
			gomponents.If(value == selected, html.Selected()),
			html.Value(fmt.Sprintf("%v", value)),
			gomponents.Textf("%v", value)))
	}
	return html.Select(
		html.Class("form-select"),
		html.Name("imageSize"),
		options,
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
				htmx.Vals(fmt.Sprintf(`{"accountID": "%s"}`, profile.AccountID)),
				htmx.Trigger("load"),
			),
		),
		Modal("recipe-details-modal", "View recipe",
			gomponents.Group{},
			gomponents.Group{
				html.Form(
					html.ID("recipe-details-form"),
				),
			},
			gomponents.Group{},
		),
	})
}
