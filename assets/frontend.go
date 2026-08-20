package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed package.json
var packageJSON []byte

type pkgManifest struct {
	Dependencies map[string]string `json:"dependencies"`
}

var deps map[string]string

func init() {
	var manifest pkgManifest
	if err := json.Unmarshal(packageJSON, &manifest); err != nil {
		panic(fmt.Sprintf("assets: parse package.json: %v", err))
	}
	deps = manifest.Dependencies
}

func version(name string) string {
	v, ok := deps[name]
	if !ok {
		panic(fmt.Sprintf("assets: missing dependency %q in package.json", name))
	}
	return v
}

func BootstrapCSS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/css/bootstrap.min.css", version("bootstrap"))
}

func BootstrapJS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/js/bootstrap.bundle.min.js", version("bootstrap"))
}

func BootstrapIconsCSS() string {
	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap-icons@%s/font/bootstrap-icons.min.css", version("bootstrap-icons"))
}

func HTMX() string {
	return fmt.Sprintf("https://unpkg.com/htmx.org@%s", version("htmx.org"))
}

func HTMXRemoveMe() string {
	return fmt.Sprintf("https://unpkg.com/htmx-ext-remove-me@%s/remove-me.js", version("htmx-ext-remove-me"))
}

func AlpineJS() string {
	return fmt.Sprintf("https://unpkg.com/alpinejs@%s/dist/cdn.min.js", version("alpinejs"))
}
