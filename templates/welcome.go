package templates

import (
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

func WelcomePage() gomponents.Node {
	return BasePage("Welcome", "/", gomponents.Group{
		html.Div(
			html.Class("m-2 p-2"),
			html.H3(html.Class("text-center"), gomponents.Text("Welcome to Grocer Recipe Hub!")),

			html.H5(html.Class("text-center"), gomponents.Text("About")),
			html.P(gomponents.Text(`
				Recipe Shopper is a quick and convenient way to create recipes using products from your local grocery store. 
				These recipes can then be browsed and quickly added to your cart for pickup, reducing the time thinking about
				meals for the week, adding items to your grocery cart, and determining how much to order.
			`)),

			html.H5(html.Class("text-center"), gomponents.Text("Grociers")),
			html.P(gomponents.Text(`
				Currently Kroger is the only supported grocer; at this time, there are no plans to support other grocers.
			`)),
			html.P(
				html.Class("fst-italic"),
				gomponents.Text(`
					Note that Recipe Shopper is unaffiliated with Kroger and makes no claim to be in partnership with, sponsored by, or endorsed by Kroger.
				`),
			),

			html.H5(html.Class("text-center"), gomponents.Text("Recipes")),
			html.P(gomponents.Text(`
				Recipe Shopper allows users to create and manage recipes that can then be quickly added to your cart for pickup; additionally, you can browse profiles
				or explore recipes others have created, and favoriting them for easy access or copying them to update to your preferences. If you'd like a recipe
				to remain private, there are visibiliy levels that can be adjusted. Instructions can be included in the recipe details so public recipes can be made
				by other users who've just discovered the recipe.
			`)),
			html.P(gomponents.Text(`
				When adding ingredients to a recipe, you can search for products sold by the integrated grocers to addto your recipe; specifying how much of the product
				the recipe requires. This is especially useful when adding recipes to your cart; if two recipes require the same prodcut, the cart adds both portions and
				rounds up to the next whole product. Alternatively, an ingredient can be marked as a staple, meaning its something that is used in small amounts and
				typically on hand. These are not included in the cart when a recipe is added, but will be included in a sepcial section that can be checked in case you
				don't have the ingredient, or need to restock.
			`)),

			html.H5(html.Class("text-center"), gomponents.Text("Cart")),
			html.P(gomponents.Text(`
				When browsing your recipes, favorites, profile recipes, or exploring recipes, you can quickly add it's ingredients to your cart. The cart temporarily tracks
				the products and their rounded quantity. Before submitting your cart to your grocer for pickup, you can update product quantities in case you have some on hand,
				you can delete unneeded products, or include staples you don't have or may be low on. When your are satisfied with your cart, you can send its contents to your
				grocer (staple ingredients that have not been included will be removed from Recipe Shopper's cart but not added to our grocer's cart). Note that you must then
				complete checkout using your grocer's online portal. 
			`)),

			html.H5(html.Class("text-center"), gomponents.Text("Profile and friends")),
			html.P(gomponents.Text(`
				You may also set a profile name. This is not required, but will prevent other users from being able to search for products created by you except by recipe name.
				If a profile name is set, then it can be found in the profiles section, which other users can then browse and find your public recipes. A profile name is required
				to access the friends feature.
			`)),
			html.P(gomponents.Text(`
				Friends: coming soon.
			`)),

			html.H5(html.Class("text-center"), gomponents.Text("Account and profile")),
			html.P(gomponents.Text(`
				It's important to select your grocer location, so that accurate product options can be provided; while not required, not setting a location will result in
				only generally available product results.
			`)),
			html.P(gomponents.Text(`
				You can also update your homepage from this welcome page to any of the recipe pages: My recipes, Favorites, or explore. Should you want to revisit this page,
				you can set your homepage back to the Welcome page.
			`)),
			html.P(gomponents.Text(`
				One final note: you may delete your account and all associated data at anytime. Any recipes, favorites, favorites others have made of your recipes profile info,
				and friends will be deleted permanently (copies others have made of your recipes will not be deleted). There is no temporary storage of your account data in case
				you return.It will be gone forever. You have been warned.
			`)),
		),
	})
}
