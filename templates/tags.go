package templates

import (
	"fmt"
	"strings"

	"github.com/densestvoid/krogerrecipeshopper/data"
	"maragu.dev/gomponents"
	htmx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/html"
)

func TagSelect() gomponents.Node {
	return html.Div(
		html.ID("tag-search-dropdown"),
		html.Class("dropdown"),
		gomponents.Attr("x-data", fmt.Sprintf("{ maxTagLen: %d }", data.MaxTagNameLength)),
		gomponents.Attr("@click.stop", `
			const searchItem = $event.target.closest('[data-tag-search-item]');
			if (searchItem) {
				const tagName = searchItem.dataset.tagName;
				const tag = document.getElementById('selected-tag-template').content.cloneNode(true);
				tag.querySelector('span.badge').innerText = tagName;
				tag.querySelector('input[name=tag]').value = tagName;
				document.getElementById('selected-tags').append(tag);
				searchItem.remove();
				$dispatch('tag-search');
				return;
			}
			const selectedItem = $event.target.closest('[data-tag-selected-item]');
			if (selectedItem) {
				selectedItem.remove();
				$dispatch('tag-search');
				return;
			}
			$dispatch('tag-search');
		`),
		gomponents.Attr("@hide-bs-dropdown.dot", "$dispatch('recipe-update')"),
		html.Button(
			html.Class("btn btn-secondary dropdown-toggle"),
			html.Role("button"),
			html.Data("bs-toggle", "dropdown"),
			html.Data("bs-auto-close", "outside"),
			gomponents.Text("Tags"),
		),
		html.Ul(
			html.Class("dropdown-menu"),
			html.Li(
				html.Class("dropdown-header"),
				html.H6(
					gomponents.Text("Search"),
				),
			),
			html.Li(
				html.Class("dropdown-item"),
				FormInput(
					"tag-search",
					"Tag search",
					nil,
					html.Input(
						html.ID("tag-search"),
						html.Class("form-control"),
						html.Type("text"),
						html.Name("search"),
						gomponents.Attr("maxlength", fmt.Sprintf("%d", data.MaxTagNameLength)),
						htmx.Get("/recipes/tags"),
						htmx.Target("#searched-tags"),
						htmx.Swap("innerHTML"),
						htmx.Include("#tag-search,#selected-tags"),
						htmx.Params("search,tag"),
						htmx.Trigger("change delay:500ms consume,input changed delay:500ms consume,tag-search from:#tag-search-dropdown"),
					),
				),
			),
			html.Li(
				html.Class("dropdown-item"),
				html.Hr(
					html.Class("dropdown-divider"),
				),
			),
			html.Template(
				html.ID("selected-tag-template"),
				TagListSelectItem(),
			),
			html.Li(
				html.Class("dropdown-header"),
				html.H6(
					gomponents.Text("Selected"),
				),
			),
			html.Li(
				html.Class("px-3"),
				html.Ul(
					html.ID("selected-tags"),
					html.Class("list-unstyled mb-0"),
				),
			),
			html.Li(
				html.Class("dropdown-item"),
				html.Hr(
					html.Class("dropdown-divider"),
				),
			),
			html.Li(
				html.Class("dropdown-header"),
				html.H6(
					gomponents.Text("Searched"),
				),
			),
			html.Li(
				html.Class("px-3"),
				html.Ul(
					html.ID("searched-tags"),
					html.Class("list-unstyled mb-0"),
				),
			),
		),
	)
}

func TagList(tags []string) gomponents.Node {
	tagItems := gomponents.Group{}
	for _, tag := range tags {
		tagItems = append(tagItems, TagListSearchItem(tag))
	}
	return tagItems
}

func TagListSelectItem() gomponents.Node {
	return html.Li(
		html.Class("dropdown-item"),
		gomponents.Attr("data-tag-selected-item", ""),
		html.Span(
			html.Class("d-inline-block badge rounded-pill text-bg-secondary me-1"),
		),
		html.Input(
			html.Type("hidden"),
			html.Name("tag"),
		),
	)
}

func TagListSearchItem(tag string) gomponents.Node {
	tag = strings.ToLower(strings.TrimSpace(tag))
	return html.Li(
		gomponents.Attr("data-tag-search-item", ""),
		gomponents.Attr("data-tag-name", tag),
		html.Span(
			html.Class("d-inline-block badge rounded-pill text-bg-secondary me-1"),
			gomponents.Text(tag),
		),
	)
}

func TagBadge(tag string) gomponents.Node {
	return html.Span(
		html.Class("d-inline-block badge rounded-pill text-bg-secondary me-1"),
		gomponents.Text(tag),
	)
}

func TagBadges(tags []string) gomponents.Node {
	badges := gomponents.Group{}
	for _, tag := range tags {
		badges = append(badges, TagBadge(tag))
	}
	return badges
}

func RecipeTagEditor(tags gomponents.Group) gomponents.Node {
	return html.Div(
		html.Class("m-2"),
		gomponents.Attr("x-data", fmt.Sprintf("{ maxTagLen: %d }", data.MaxTagNameLength)),
		gomponents.Attr("@click", `
			if ($event.target.closest('[data-recipe-tag-add]')) {
				const tag = document.getElementById('recipe-tag').content.cloneNode(true);
				const span = tag.querySelector('.tag-badge');
				document.getElementById('recipe-tags').append(tag);
				span.focus();
			}
			if ($event.target.closest('[data-recipe-tag-remove]')) {
				$event.target.closest('.recipe-tag').remove();
			}
		`),
		gomponents.Attr("@focusout", `
			const badge = $event.target.closest('.tag-badge');
			if (!badge || !$event.currentTarget.contains(badge)) return;
			const root = badge.closest('.recipe-tag');
			const input = root.querySelector('input[name=tag]');
			let tag = badge.innerText.trim().slice(0, maxTagLen).toLowerCase();
			if (!tag) {
				root.remove();
			} else {
				badge.innerText = tag;
				input.value = tag;
			}
		`),
		gomponents.Attr("@keydown.enter.prevent", "$event.target.blur()"),
		html.Template(
			html.ID("recipe-tag"),
			Tag(""),
		),
		html.Label(gomponents.Text("Tags:")),
		html.Span(
			html.Class("d-inline-block badge rounded-pill text-bg-primary mx-2"),
			gomponents.Attr("data-recipe-tag-add", ""),
			html.I(html.Class("bi bi-plus-circle")),
		),
		html.Div(
			html.ID("recipe-tags"),
			tags,
		),
	)
}

func Tag(tag string) gomponents.Node {
	tag = strings.ToLower(strings.TrimSpace(tag))
	return html.Div(
		html.Class("d-inline-block badge rounded-pill text-bg-secondary mx-1 recipe-tag"),
		html.Span(
			html.Class("tag-badge me-1"),
			gomponents.Attr("contentEditable"),
			gomponents.Text(tag),
		),
		html.I(
			html.Class("bi bi-x-circle ms-1"),
			gomponents.Attr("data-recipe-tag-remove", ""),
		),
		html.Input(
			html.Type("hidden"),
			html.Name("tag"),
			html.Value(tag),
		),
	)
}
