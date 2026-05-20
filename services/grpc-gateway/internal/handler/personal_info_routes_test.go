package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPersonalInfoRoutes_RejectsUnknownMethod(t *testing.T) {
	h := &AuthHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/personal-info", nil)

	PersonalInfoRoutes(h)(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestPersonalInfoRoutes_AcceptsSpoofedPatch(t *testing.T) {
	h := &AuthHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/personal-info?_method=patch", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	PersonalInfoRoutes(h)(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Fatal("POST with _method=patch must reach UpdatePersonalInfo, not 404")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (no auth context in test)", rr.Code, http.StatusUnauthorized)
	}
}
