package service

import "testing"

func TestResolveSMSAPIKeyPrefersValidSMSKey(t *testing.T) {
	t.Setenv("SMS_API_KEY", "real-key")
	t.Setenv("KAVENEGAR_API_KEY", "fallback-key")

	if got := ResolveSMSAPIKey(); got != "real-key" {
		t.Fatalf("expected real-key, got %q", got)
	}
}

func TestResolveSMSAPIKeyFallsBackFromPlaceholder(t *testing.T) {
	t.Setenv("SMS_API_KEY", "change-me")
	t.Setenv("KAVENEGAR_API_KEY", "fallback-key")

	if got := ResolveSMSAPIKey(); got != "fallback-key" {
		t.Fatalf("expected fallback-key, got %q", got)
	}
}

func TestSMSAPIKeySource(t *testing.T) {
	t.Setenv("SMS_API_KEY", "real-key")
	t.Setenv("KAVENEGAR_API_KEY", "other-key")
	if got := SMSAPIKeySource(); got != "SMS_API_KEY" {
		t.Fatalf("expected SMS_API_KEY, got %q", got)
	}
}

func TestIsPlaceholderSMSAPIKey(t *testing.T) {
	if !IsPlaceholderSMSAPIKey("change-me") {
		t.Fatal("expected change-me to be a placeholder")
	}
	if IsPlaceholderSMSAPIKey("693835337A3377547771646A327733396D6D79393539744E6A5372487644456F3448434C773974337234733D") {
		t.Fatal("expected real key not to be a placeholder")
	}
}
