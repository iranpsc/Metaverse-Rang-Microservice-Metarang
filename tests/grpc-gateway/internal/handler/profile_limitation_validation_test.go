package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metarang/grpc-gateway/internal/handler"
	"metarang/grpc-gateway/internal/testutil"
)

func validOptionsJSON() map[string]bool {
	return map[string]bool{
		"follow":                  false,
		"send_message":            false,
		"share":                   true,
		"send_ticket":             true,
		"view_profile_images":     false,
		"view_features_locations": true,
	}
}

func TestCreateProfileLimitation_Validation(t *testing.T) {
	h := &handler.AuthHandler{}

	t.Run("missing options", func(t *testing.T) {
		body := `{"limited_user_id": 2}`
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", strings.NewReader(body))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "options")
	})

	t.Run("each missing option key", func(t *testing.T) {
		for _, missing := range []string{
			"follow", "send_message", "share", "send_ticket", "view_profile_images", "view_features_locations",
		} {
			opts := validOptionsJSON()
			delete(opts, missing)
			payload := map[string]interface{}{
				"limited_user_id": 2,
				"options":         opts,
			}
			raw, _ := json.Marshal(payload)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
			req = testutil.RequestWithUser(req, 1)
			h.CreateProfileLimitation(rr, req)
			if rr.Code != http.StatusUnprocessableEntity {
				t.Fatalf("missing %s: status = %d, want 422; body=%s", missing, rr.Code, rr.Body.String())
			}
			assertFieldError(t, rr.Body.Bytes(), "options."+missing)
		}
	})

	t.Run("unknown option key", func(t *testing.T) {
		opts := validOptionsJSON()
		payload := map[string]interface{}{
			"limited_user_id": 2,
			"options": map[string]interface{}{
				"follow":                  opts["follow"],
				"send_message":            opts["send_message"],
				"share":                   opts["share"],
				"send_ticket":             opts["send_ticket"],
				"view_profile_images":     opts["view_profile_images"],
				"view_features_locations": opts["view_features_locations"],
				"extra":                   true,
			},
		}
		raw, _ := json.Marshal(payload)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "options.extra")
	})

	t.Run("non-boolean option value", func(t *testing.T) {
		payload := map[string]interface{}{
			"limited_user_id": 2,
			"options": map[string]interface{}{
				"follow":                  "no",
				"send_message":            false,
				"share":                   true,
				"send_ticket":             true,
				"view_profile_images":     false,
				"view_features_locations": true,
			},
		}
		raw, _ := json.Marshal(payload)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "options.follow")
	})

	t.Run("oversized note", func(t *testing.T) {
		note := strings.Repeat("a", 501)
		payload := map[string]interface{}{
			"limited_user_id": 2,
			"options":         validOptionsJSON(),
			"note":            note,
		}
		raw, _ := json.Marshal(payload)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "note")
	})

	t.Run("missing limited_user_id", func(t *testing.T) {
		payload := map[string]interface{}{"options": validOptionsJSON()}
		raw, _ := json.Marshal(payload)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "limited_user_id")
	})

	t.Run("zero limited_user_id", func(t *testing.T) {
		payload := map[string]interface{}{
			"limited_user_id": 0,
			"options":         validOptionsJSON(),
		}
		raw, _ := json.Marshal(payload)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", bytes.NewReader(raw))
		req = testutil.RequestWithUser(req, 1)
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
		}
		assertFieldError(t, rr.Body.Bytes(), "limited_user_id")
	})

	t.Run("unauthenticated", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profile-limitations", strings.NewReader(`{}`))
		h.CreateProfileLimitation(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rr.Code)
		}
	})
}

func TestUpdateProfileLimitation_RequiresOptions(t *testing.T) {
	h := &handler.AuthHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/profile-limitations/1", strings.NewReader(`{"note":"x"}`))
	req = testutil.RequestWithUser(req, 1)
	h.UpdateProfileLimitation(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
	}
	assertFieldError(t, rr.Body.Bytes(), "options")
}

func TestGetProfileLimitations_RequiresAuth(t *testing.T) {
	h := &handler.AuthHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/2/profile-limitations", nil)
	h.GetProfileLimitations(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func assertFieldError(t *testing.T, body []byte, field string) {
	t.Helper()
	var resp struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v; body=%s", err, string(body))
	}
	if _, ok := resp.Errors[field]; !ok {
		t.Fatalf("expected field error %q in %v", field, resp.Errors)
	}
}
