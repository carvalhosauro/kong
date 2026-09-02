package schedule

import (
	"testing"
	"time"
)

func TestNextRun_ORRuleDayAndWeekday(t *testing.T) {
	// Vixie OR rule: restricted DOM and DOW combine with OR.
	// 0 0 13 * 5 = midnight on day 13 OR Friday.
	loc := time.UTC
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, loc) // Thursday

	next, err := NextRun("0 0 13 * 5", from, loc)
	if err != nil {
		t.Fatalf("NextRun: %v", err)
	}
	if next == nil {
		t.Fatal("expected next run")
	}

	// Jan 2 is Friday — should fire before Jan 13.
	expected := time.Date(2026, 1, 2, 0, 0, 0, 0, loc)
	if !next.Equal(expected) {
		t.Fatalf("got %v, want %v", next, expected)
	}
}

func TestNextRun_Reboot(t *testing.T) {
	next, err := NextRun("@reboot", time.Now(), time.UTC)
	if err != nil {
		t.Fatalf("NextRun: %v", err)
	}
	if next != nil {
		t.Fatalf("expected nil for @reboot, got %v", next)
	}
}

func TestTruncateCommand(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`echo hello`, `echo hello`},
		{`echo hello%world`, `echo hello`},
		{`echo \%literal`, `echo \%literal`},
	}
	for _, tc := range tests {
		got := TruncateCommand(tc.in)
		if got != tc.want {
			t.Errorf("TruncateCommand(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeMacros(t *testing.T) {
	norm, reboot, err := Normalize("@daily")
	if err != nil || reboot || norm != "0 0 * * *" {
		t.Fatalf("@daily: norm=%q reboot=%v err=%v", norm, reboot, err)
	}
	_, reboot, err = Normalize("@reboot")
	if err != nil || !reboot {
		t.Fatalf("@reboot: reboot=%v err=%v", reboot, err)
	}
}
