package handler

import (
	"testing"

	pb "metargb/shared/pb/auth"
)

func TestParseBankAccountErrors(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		if got := parseBankAccountErrors(""); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("json array", func(t *testing.T) {
		got := parseBankAccountErrors(`["invalid sheba","invalid card"]`)
		arr, ok := got.([]interface{})
		if !ok || len(arr) != 2 {
			t.Fatalf("expected array with 2 items, got %#v", got)
		}
	})

	t.Run("plain string fallback", func(t *testing.T) {
		got := parseBankAccountErrors("not-json")
		if got != "not-json" {
			t.Fatalf("expected raw string, got %#v", got)
		}
	})
}

func TestFormatBankAccountResource(t *testing.T) {
	resp := &pb.BankAccountResponse{
		Id:       1,
		BankName: "Tejarat",
		ShabaNum: "IR820540102680020817909002",
		CardNum:  "6037997551234567",
		Status:   -1,
		Errors:   `["rejected reason"]`,
	}

	formatted := formatBankAccountResource(resp)
	if formatted["id"] != uint64(1) {
		t.Fatalf("unexpected id: %#v", formatted["id"])
	}
	errors, ok := formatted["errors"].([]interface{})
	if !ok || len(errors) != 1 {
		t.Fatalf("expected parsed errors array, got %#v", formatted["errors"])
	}
	if _, ok := formatted["errors"].(string); ok {
		t.Fatal("errors should not be returned as a raw string")
	}
}
