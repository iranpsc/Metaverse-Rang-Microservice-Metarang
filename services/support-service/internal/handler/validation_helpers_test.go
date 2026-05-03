package handler

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		err  error
		want codes.Code
	}{
		{nil, codes.OK},
		{errors.New("Unauthorized access"), codes.PermissionDenied},
		{errors.New("resource not found"), codes.NotFound},
		{errors.New("cannot respond to closed ticket"), codes.FailedPrecondition},
		{errors.New("ticket is already closed"), codes.FailedPrecondition},
		{errors.New("db exploded"), codes.Internal},
	}
	for _, tc := range tests {
		got := MapServiceError(tc.err)
		if tc.err == nil {
			if got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			continue
		}
		st, ok := status.FromError(got)
		if !ok || st.Code() != tc.want {
			t.Fatalf("err=%v want=%v got=%v", tc.err, tc.want, got)
		}
	}
}

func TestValidateMaxLen(t *testing.T) {
	if m := validateMaxLen("f", "abc", 2, "en"); m == nil {
		t.Fatal("expected error")
	}
	if m := validateMaxLen("f", "ab", 5, "en"); len(m) != 0 {
		t.Fatalf("unexpected %v", m)
	}
}

func TestValidateReportSubject_EmptyUsesRequired(t *testing.T) {
	m := validateReportSubject("", "en")
	if m == nil || m["subject"] == "" {
		t.Fatalf("unexpected %v", m)
	}
}
