package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	pb "metarang/shared/pb/auth"
)

func TestParseCreateProfileLimitationBody_NotePresence(t *testing.T) {
	validBodyPrefix := `{
			"limited_user_id": 2,
			"options": {
				"follow": false,
				"send_message": false,
				"share": true,
				"send_ticket": true,
				"view_profile_images": false,
				"view_features_locations": true
			}`

	t.Run("omitted note", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(validBodyPrefix+`}`))
		input, errs := parseCreateProfileLimitationBody(req)
		if errs != nil {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if input.Note.Present {
			t.Fatal("note should be omitted")
		}
		if notePtrFromInput(input.Note) != nil {
			t.Fatal("proto note should be unset when omitted")
		}
	})

	t.Run("explicit null note", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(validBodyPrefix+`,"note": null}`))
		input, errs := parseCreateProfileLimitationBody(req)
		if errs != nil {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !input.Note.Present || input.Note.Value != nil {
			t.Fatalf("expected explicit clear note, got %+v", input.Note)
		}
		ptr := notePtrFromInput(input.Note)
		if ptr == nil || *ptr != "" {
			t.Fatalf("clear note should wire as empty optional string, got %v", ptr)
		}
	})

	t.Run("string note", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(validBodyPrefix+`,"note": "hello"}`))
		input, errs := parseCreateProfileLimitationBody(req)
		if errs != nil {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !input.Note.Present || input.Note.Value == nil || *input.Note.Value != "hello" {
			t.Fatalf("expected string note, got %+v", input.Note)
		}
	})
}

func TestParseUpdateProfileLimitationBody_OmitVsClearNote(t *testing.T) {
	optionsJSON := `{
		"follow": true,
		"send_message": true,
		"share": true,
		"send_ticket": true,
		"view_profile_images": true,
		"view_features_locations": true
	}`

	t.Run("omitted note", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/", strings.NewReader(`{"options":`+optionsJSON+`}`))
		input, errs := parseUpdateProfileLimitationBody(req)
		if errs != nil {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if input.Note.Present {
			t.Fatal("note should be omitted")
		}
	})

	t.Run("cleared note", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/", strings.NewReader(`{"options":`+optionsJSON+`,"note":null}`))
		input, errs := parseUpdateProfileLimitationBody(req)
		if errs != nil {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !input.Note.Present || input.Note.Value != nil {
			t.Fatalf("expected cleared note, got %+v", input.Note)
		}
	})
}

func TestProfileLimitationResourceJSON(t *testing.T) {
	follow := false
	trueVal := true
	note := "secret"
	data := &pb.ProfileLimitation{
		Id:            7,
		LimiterUserId: 42,
		LimitedUserId: 1234,
		Options: &pb.ProfileLimitationOptions{
			Follow:                &follow,
			SendMessage:           &trueVal,
			Share:                 &trueVal,
			SendTicket:            &trueVal,
			ViewProfileImages:     &follow,
			ViewFeaturesLocations: &trueVal,
		},
		Note: &note,
	}

	asLimiter := profileLimitationResourceJSON(data, 42)
	if _, ok := asLimiter["created_at"]; ok {
		t.Fatal("created_at must not be present")
	}
	if _, ok := asLimiter["updated_at"]; ok {
		t.Fatal("updated_at must not be present")
	}
	if asLimiter["note"] != "secret" {
		t.Fatalf("limiter should see note, got %v", asLimiter["note"])
	}

	empty := ""
	data.Note = &empty
	asLimiterNull := profileLimitationResourceJSON(data, 42)
	if asLimiterNull["note"] != nil {
		t.Fatalf("limiter should see null note, got %v", asLimiterNull["note"])
	}

	asLimited := profileLimitationResourceJSON(data, 1234)
	if _, ok := asLimited["note"]; ok {
		t.Fatal("limited user must not see note key")
	}
}
