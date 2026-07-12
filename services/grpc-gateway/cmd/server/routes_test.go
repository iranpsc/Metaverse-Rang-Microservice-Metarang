package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterExactAndTrailingSlash(t *testing.T) {
	mux := http.NewServeMux()
	registerExactAndTrailingSlash(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "/api/order/callback")

	for _, path := range []string{"/api/order/callback", "/api/order/callback/"} {
		req := httptest.NewRequest(http.MethodGet, path+"?order_id=1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, rec.Code)
		}
	}
}
