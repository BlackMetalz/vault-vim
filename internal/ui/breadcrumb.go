package ui

import "strings"

func renderBreadcrumb(parts []string) string {
	var rendered []string
	for _, p := range parts {
		rendered = append(rendered, breadcrumbStyle.Render(p))
	}
	return strings.Join(rendered, breadcrumbSep.String())
}
