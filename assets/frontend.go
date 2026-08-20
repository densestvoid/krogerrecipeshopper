package assets

import (
	_ "embed"
	"encoding/json/v2"
	"fmt"
)

//go:embed package.json
var packageJSON []byte

type frontendDependencies struct {
	Alpinejs        string `json:"alpinejs"`
	Bootstrap       string `json:"bootstrap"`
	BootstrapIcons  string `json:"bootstrap-icons"`
	HTMXOrg         string `json:"htmx.org"`
	HTMXExtRemoveMe string `json:"htmx-ext-remove-me"`
}

var frontend frontendDependencies

func init() {
	var manifest struct {
		Dependencies frontendDependencies `json:"dependencies"`
	}
	if err := json.Unmarshal(packageJSON, &manifest); err != nil {
		panic(fmt.Sprintf("assets: parse package.json: %v", err))
	}
	frontend = manifest.Dependencies
}

func BootstrapCSS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/css/bootstrap.min.css", frontend.Bootstrap)
}

func BootstrapJS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/js/bootstrap.bundle.min.js", frontend.Bootstrap)
}

func BootstrapIconsCSS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap-icons@%s/font/bootstrap-icons.min.css", frontend.BootstrapIcons)
}

func HTMX() string {
	return fmt.Sprintf("https://unpkg.com/htmx.org@%s", frontend.HTMXOrg)
}

func HTMXRemoveMe() string {
	return fmt.Sprintf("https://unpkg.com/htmx-ext-remove-me@%s/remove-me.js", frontend.HTMXExtRemoveMe)
}

func AlpineJS() string {
	return fmt.Sprintf("https://unpkg.com/alpinejs@%s/dist/cdn.min.js", frontend.Alpinejs)
}
