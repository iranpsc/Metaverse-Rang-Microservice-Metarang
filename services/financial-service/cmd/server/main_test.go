package main

import (
	"os"
	"testing"
)

func TestResolveSadadCallbackURL(t *testing.T) {
	t.Setenv("SADAD_CALLBACK_URL", "")
	t.Setenv("PAYMENT_CALLBACK_URL", "")
	t.Setenv("PROJECT_URL", "http://localhost:8000")

	if got := resolveSadadCallbackURL(); got != "http://localhost:8000/api/payment/callback" {
		t.Fatalf("expected default callback URL, got %q", got)
	}

	t.Setenv("SADAD_CALLBACK_URL", "${PROJECT_URL}/api/payment/callback")
	if got := resolveSadadCallbackURL(); got != "http://localhost:8000/api/payment/callback" {
		t.Fatalf("expected expanded callback URL, got %q", got)
	}

	t.Setenv("SADAD_CALLBACK_URL", "${FRONTEND_URL}/payment/verify")
	t.Setenv("FRONTEND_URL", "http://localhost:5173")
	if got := resolveSadadCallbackURL(); got != "http://localhost:8000/api/payment/callback" {
		t.Fatalf("expected fallback when callback URL points to frontend verify page, got %q", got)
	}

	t.Setenv("SADAD_CALLBACK_URL", "https://api.example.com/api/payment/callback")
	if got := resolveSadadCallbackURL(); got != "https://api.example.com/api/payment/callback" {
		t.Fatalf("expected explicit callback URL, got %q", got)
	}
}

func TestNormalizePaymentCallbackURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantOK  bool
	}{
		{
			name:   "valid callback url",
			raw:    "http://localhost:8000/api/payment/callback",
			want:   "http://localhost:8000/api/payment/callback",
			wantOK: true,
		},
		{
			name:   "base url without path",
			raw:    "http://localhost:8000",
			want:   "http://localhost:8000/api/payment/callback",
			wantOK: true,
		},
		{
			name:   "frontend verify page rejected",
			raw:    "http://localhost:5173/payment/verify",
			wantOK: false,
		},
		{
			name:   "unrelated path rejected",
			raw:    "http://localhost:8000/api/order",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizePaymentCallbackURL(tt.raw)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if ok && got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
