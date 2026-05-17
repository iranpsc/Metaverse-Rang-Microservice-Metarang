package handler

import (
	"context"
	"os"
	"strings"

	"google.golang.org/grpc/metadata"
)

var projectLocale string

// SetProjectLocale sets the project locale for all training-service handlers.
func SetProjectLocale(locale string) {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale != "fa" && locale != "en" {
		locale = "en"
	}
	projectLocale = locale
}

func getLocale(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("locale"); len(vals) > 0 && vals[0] != "" {
			locale := strings.ToLower(strings.TrimSpace(vals[0]))
			if locale == "fa" || locale == "en" {
				return locale
			}
		}
		if vals := md.Get("accept-language"); len(vals) > 0 {
			lang := strings.ToLower(strings.TrimSpace(strings.Split(vals[0], ",")[0]))
			if strings.HasPrefix(lang, "fa") {
				return "fa"
			}
		}
	}
	if projectLocale != "" {
		return projectLocale
	}
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("PROJECT_LOCALE"))); env == "fa" {
		return "fa"
	}
	return "en"
}
