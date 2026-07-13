package jalali_test

import (
	"testing"
	"time"

	"metarang/shared/pkg/jalali"
)

func TestJalaliRoundTrip(t *testing.T) {
	cases := []struct {
		gregorian string
		jalali    string
	}{
		{"2026-07-06", "1405/07/01"},
		{"2026-09-22", "1405/09/19"},
	}

	for _, tc := range cases {
		gt, err := time.Parse("2006-01-02", tc.gregorian)
		if err != nil {
			t.Fatal(err)
		}
		if got := jalali.CarbonToJalali(gt); got != tc.jalali {
			t.Fatalf("CarbonToJalali(%s) = %s, want %s", tc.gregorian, got, tc.jalali)
		}

		back, err := jalali.JalaliToCarbon(tc.jalali)
		if err != nil {
			t.Fatalf("JalaliToCarbon(%s): %v", tc.jalali, err)
		}
		if back.Format("2006-01-02") != tc.gregorian {
			t.Fatalf("JalaliToCarbon(%s) = %s, want %s", tc.jalali, back.Format("2006-01-02"), tc.gregorian)
		}
	}
}

func TestEvent754DateFilter(t *testing.T) {
	filter, err := jalali.JalaliToCarbon("1405/07/15")
	if err != nil {
		t.Fatal(err)
	}
	startsAt := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)

	filterDate := time.Date(filter.Year(), filter.Month(), filter.Day(), 0, 0, 0, 0, time.UTC)
	startDate := time.Date(startsAt.Year(), startsAt.Month(), startsAt.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(endsAt.Year(), endsAt.Month(), endsAt.Day(), 0, 0, 0, 0, time.UTC)

	if startDate.After(filterDate) || endDate.Before(filterDate) {
		t.Fatalf("filter %s not within [%s, %s]", filterDate.Format("2006-01-02"), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}
}
