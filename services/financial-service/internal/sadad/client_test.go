package sadad

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSadadLocalDateTimeUsesTehranTimezone(t *testing.T) {
	tehran, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Fatalf("failed to load Tehran location: %v", err)
	}

	now := time.Now().In(tehran)
	got := sadadLocalDateTime()

	if !strings.Contains(got, now.Format("2006")) {
		t.Fatalf("expected year %s in LocalDateTime, got %q", now.Format("2006"), got)
	}
	if strings.Contains(got, "-") {
		t.Fatalf("expected Sadad date format without dashes, got %q", got)
	}
}

func TestRequestPaymentSendsPaymentIdentityAndLocalDateTime(t *testing.T) {
	var received paymentByIdentityRequestBody
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != "" {
			t.Fatalf("expected empty User-Agent, got %q", ua)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ResCode": 0,
			"Token":   "test-token",
		})
	}))
	defer server.Close()

	client := NewClientWithEndpoints(Endpoints{
		PaymentRequestURL: server.URL,
		VerifyURL:         server.URL,
		GatewayURL:        "https://example.com/purchase",
		Multiplexed:       true,
	})
	resp, err := client.RequestPayment(RequestParams{
		MerchantID:      "merchant",
		TerminalID:      "terminal",
		TransactionKey:  "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0",
		OrderID:         "42",
		Amount:          1000,
		ReturnURL:       "https://example.com/callback",
		PaymentIdentity: "identity-123",
	})
	if err != nil {
		t.Fatalf("RequestPayment failed: %v", err)
	}
	if !resp.Success() {
		t.Fatalf("expected success response, got ResCode=%q", resp.ResCode)
	}
	if received.PaymentIdentity != "identity-123" {
		t.Fatalf("expected PaymentIdentity in request, got %q", received.PaymentIdentity)
	}
	if received.LocalDateTime == "" {
		t.Fatal("expected LocalDateTime in request")
	}
}

func TestSandboxRequestPaymentOmitsPaymentIdentity(t *testing.T) {
	var received paymentRequestBody
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ResCode": 0,
			"Token":   "sandbox-token",
		})
	}))
	defer server.Close()

	client := NewClientWithEndpoints(Endpoints{
		PaymentRequestURL: server.URL,
		VerifyURL:         server.URL,
		GatewayURL:        SandboxEndpoints.GatewayURL,
		Multiplexed:       false,
	})
	resp, err := client.RequestPayment(RequestParams{
		MerchantID:     "46645",
		TerminalID:     "GBHDTY98",
		TransactionKey: "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0",
		OrderID:        "1",
		Amount:         10000,
		ReturnURL:      "http://localhost/callback",
	})
	if err != nil {
		t.Fatalf("RequestPayment failed: %v", err)
	}
	if !resp.Success() {
		t.Fatalf("expected success response, got ResCode=%q", resp.ResCode)
	}
	if received.MerchantID != "46645" || received.TerminalID != "GBHDTY98" {
		t.Fatalf("unexpected request body: %+v", received)
	}
	wantURL := SandboxEndpoints.GatewayURL + "?Token=sandbox-token"
	if got := resp.URL(); got != wantURL {
		t.Fatalf("expected %q, got %q", wantURL, got)
	}
}
