package handler

import (
	"testing"
)

func TestValidateRequired_StringEmpty(t *testing.T) {
	errs := validateRequired("field", "", "en")
	if errs["field"] == "" {
		t.Fatal("expected error message")
	}
}

func TestValidateRequired_StringOK(t *testing.T) {
	errs := validateRequired("field", "x", "en")
	if len(errs) != 0 {
		t.Fatal(errs)
	}
}

func TestValidateRequired_FAPLocale(t *testing.T) {
	errs := validateRequired("field", "", "fa")
	if errs["field"] == "" {
		t.Fatal("expected fa message")
	}
}
