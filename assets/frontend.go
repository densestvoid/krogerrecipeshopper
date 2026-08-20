package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed package.json
var packageJSON []byte

type packageManifest struct {
	Dependencies frontendDependencies `json:"dependencies"`
}

type frontendDependencies struct {
	Bootstrap       string `json:"bootstrap"`
	BootstrapIcons  string `json:"bootstrap-icons"`
	HTMXOrg         string `json:"htmx.org"`
	HTMXExtRemoveMe string `json:"htmx-ext-remove-me"`
}

var frontend frontendDependencies

func init() {
	var manifest packageManifest
	if err := json.Unmarshal(packageJSON, &manifest); err != nil {
		panic(fmt.Sprintf("assets: parse package.json: %v", err))
	}
	frontend = manifest.Dependencies
	if err := frontend.validate(); err != nil {
		panic(fmt.Sprintf("assets: invalid package.json: %v", err))
	}
}

func (d frontendDependencies) validate() error {
	if d.Bootstrap == "" {
		return fmt.Errorf("missing bootstrap")
	}
	if d.BootstrapIcons == "" {
		return fmt.Errorf("missing bootstrap-icons")
	}
	if d.HTMXOrg == "" {
		return fmt.Errorf("missing htmx.org")
	}
	if d.HTMXExtRemoveMe == "" {
		return fmt.Errorf("missing htmx-ext-remove-me")
	}
	return nil
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

// AlpineJS was previously unpinned on unpkg; not in package.json until Dependabot adds it.
func AlpineJS() string {
	return "https://unpkg.com/alpinejs"
}
