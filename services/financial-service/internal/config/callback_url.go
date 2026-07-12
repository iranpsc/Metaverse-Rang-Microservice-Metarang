package config

import (
	"log"
	"net/url"
	"os"
	"strings"
)

const (
	OrderCallbackPath         = "/api/order/callback"
	legacyPaymentCallbackPath = "/api/payment/callback"
	defaultProjectURL         = "http://localhost:8000"
)

// ResolveSadadCallbackURL returns the Sadad ReturnUrl base (without order_id query param).
// The gateway must redirect users to the API callback endpoint, never the frontend verify page.
// Supports ${PROJECT_URL} expansion in config.env (e.g. SADAD_CALLBACK_URL=${PROJECT_URL}/api/order/callback).
func ResolveSadadCallbackURL() string {
	for _, key := range []string{"SADAD_CALLBACK_URL", "PAYMENT_CALLBACK_URL"} {
		if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
			if expanded := strings.TrimSpace(os.ExpandEnv(raw)); expanded != "" {
				if normalized, ok := NormalizePaymentCallbackURL(expanded); ok {
					return normalized
				}
				log.Printf("Warning: %s=%q is not a valid API callback URL; falling back to PROJECT_URL", key, expanded)
			}
		}
	}

	projectURL := strings.TrimSpace(os.ExpandEnv(getEnv("PROJECT_URL", defaultProjectURL)))
	return strings.TrimSuffix(projectURL, "/") + OrderCallbackPath
}

// NormalizePaymentCallbackURL validates and normalizes Sadad callback URLs.
func NormalizePaymentCallbackURL(raw string) (string, bool) {
	if strings.Contains(raw, "/payment/verify") {
		return "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	switch path {
	case OrderCallbackPath:
		return strings.TrimSuffix(raw, "/"), true
	case legacyPaymentCallbackPath:
		parsed.Path = OrderCallbackPath
		return strings.TrimSuffix(parsed.String(), "/"), true
	}

	if path == "" || path == "/" {
		parsed.Path = OrderCallbackPath
		return strings.TrimSuffix(parsed.String(), "/"), true
	}

	return "", false
}

// ResolveProjectURL returns the normalized project base URL from PROJECT_URL.
func ResolveProjectURL() string {
	raw := strings.TrimSpace(getEnv("PROJECT_URL", defaultProjectURL))
	expanded := strings.TrimSpace(os.ExpandEnv(raw))
	return strings.TrimSuffix(normalizeURLScheme(expanded), "/")
}

// ResolveFrontendURL returns the normalized frontend base URL from FRONTEND_URL.
func ResolveFrontendURL() string {
	raw := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if raw == "" {
		return ""
	}
	expanded := strings.TrimSpace(os.ExpandEnv(raw))
	return strings.TrimSuffix(normalizeURLScheme(expanded), "/")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func normalizeURLScheme(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	if strings.Contains(rawURL, "://") {
		return rawURL
	}
	return "http://" + rawURL
}
