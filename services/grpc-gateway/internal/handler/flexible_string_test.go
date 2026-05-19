package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestFlexibleString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "string", input: `"123456"`, want: "123456"},
		{name: "integer", input: `123456`, want: "123456"},
		{name: "persian string", input: `"۱۲۳۴۵۶"`, want: "123456"},
		{name: "null", input: `null`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got flexibleString
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestDecodeRequestBody_AccountSecurityVerifyCode(t *testing.T) {
	t.Run("json numeric code", func(t *testing.T) {
		body := []byte(`{"code":123456}`)
		req, err := http.NewRequest(http.MethodPost, "/api/account/security/verify", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		var parsed struct {
			Code flexibleString `json:"code" form:"code"`
		}
		if err := decodeRequestBody(req, &parsed); err != nil {
			t.Fatalf("decodeRequestBody: %v", err)
		}
		if parsed.Code.String() != "123456" {
			t.Fatalf("got code %q, want 123456", parsed.Code.String())
		}
	})

	t.Run("form-urlencoded code", func(t *testing.T) {
		body := []byte("code=654321")
		req, err := http.NewRequest(http.MethodPost, "/api/account/security/verify", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ContentLength = int64(len(body))

		var parsed struct {
			Code flexibleString `json:"code" form:"code"`
		}
		if err := decodeRequestBody(req, &parsed); err != nil {
			t.Fatalf("decodeRequestBody: %v", err)
		}
		if parsed.Code.String() != "654321" {
			t.Fatalf("got code %q, want 654321", parsed.Code.String())
		}
	})

	t.Run("invalid json without query params returns error", func(t *testing.T) {
		body := []byte(`{"code":}`)
		req, err := http.NewRequest(http.MethodPost, "/api/account/security/verify", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		var parsed struct {
			Code flexibleString `json:"code" form:"code"`
		}
		if err := decodeRequestBody(req, &parsed); err == nil {
			t.Fatal("expected decode error for malformed JSON without query params")
		}
	})
}
