package utils

import (
	"strings"
	"testing"
	"time"
)

func TestFormatJalaliDate(t *testing.T) {
	ts := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	s := FormatJalaliDate(ts)
	if !strings.Contains(s, "/") || len(s) < 8 {
		t.Fatalf("unexpected %q", s)
	}
}

func TestFormatJalaliTime(t *testing.T) {
	ts := time.Date(2024, 1, 1, 9, 5, 3, 0, time.UTC)
	if got := FormatJalaliTime(ts); got != "09:05:03" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatJalaliDateTime(t *testing.T) {
	ts := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	s := FormatJalaliDateTime(ts)
	if !strings.Contains(s, "14:30:00") {
		t.Fatalf("unexpected %q", s)
	}
}

func TestGregorianToJalali_Branches(t *testing.T) {
	// Exercise year <= 1600 branch
	j := GregorianToJalali(time.Date(900, 3, 1, 0, 0, 0, 0, time.UTC))
	if j.Year <= 0 || j.Month < 1 || j.Day < 1 {
		t.Fatalf("unexpected %+v", j)
	}
	// Exercise gm > 2 leap branch
	j2 := GregorianToJalali(time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC))
	if j2.Month < 1 {
		t.Fatalf("unexpected %+v", j2)
	}
}
