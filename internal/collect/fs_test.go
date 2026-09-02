package collect

import (
	"path/filepath"
	"testing"
)

func TestJoinFS_EmptyRootUsesAbsolutePaths(t *testing.T) {
	got := joinFS("", "etc", "crontab")
	want := filepath.Join(string(filepath.Separator), "etc", "crontab")
	if got != want {
		t.Fatalf("joinFS empty root = %q, want %q", got, want)
	}
}

func TestJoinFS_FixtureRoot(t *testing.T) {
	got := joinFS("/fixtures", "etc", "crontab")
	want := filepath.Join("/fixtures", "etc", "crontab")
	if got != want {
		t.Fatalf("joinFS fixture = %q, want %q", got, want)
	}
}
