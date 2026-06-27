package sadad

import "testing"

func TestSandboxEndpointsMatchBankTestURLs(t *testing.T) {
	cases := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "payment request",
			got:      SandboxEndpoints.PaymentRequestURL,
			expected: "https://sandbox.banktest.ir/melli/sadad.shaparak.ir/VPG/api/v0/Request/PaymentRequest",
		},
		{
			name:     "verify",
			got:      SandboxEndpoints.VerifyURL,
			expected: "https://sandbox.banktest.ir/melli/sadad.shaparak.ir/VPG/api/v0/Advice/Verify",
		},
		{
			name:     "purchase gateway",
			got:      SandboxEndpoints.GatewayURL,
			expected: "https://sandbox.banktest.ir/melli/sadad.shaparak.ir/VPG/Purchase",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, tc.got)
			}
		})
	}

	if SandboxEndpoints.Multiplexed {
		t.Fatal("sandbox endpoints must not use PaymentByIdentity multiplexing")
	}
}

func TestRequestResponseURLUsesClientGateway(t *testing.T) {
	client := NewClientWithSandbox(true)
	resp := &RequestResponse{
		ResCode:    "0",
		Token:      "test-token",
		gatewayURL: client.endpoints.GatewayURL,
	}

	want := SandboxEndpoints.GatewayURL + "?Token=test-token"
	if got := resp.URL(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
