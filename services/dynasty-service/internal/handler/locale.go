package handler

import (
	"context"
	"strings"
)

var projectLocale string

// SetProjectLocale sets global locale for handler validation messages.
func SetProjectLocale(locale string) {
	locale = normalizeLocale(locale)
	projectLocale = locale
}

func getLocale(ctx context.Context) string {
	_ = ctx
	if projectLocale == "" {
		return "en"
	}
	return projectLocale
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "fa" {
		return "fa"
	}
	return "en"
}
