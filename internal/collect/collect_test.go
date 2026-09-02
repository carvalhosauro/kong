package collect

import (
	"path/filepath"
	"testing"

	"github.com/carvalhosauro/kong/internal/model"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "fixtures", "debian-12")
}

func TestCollectDebianFixture(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(result.Jobs) == 0 {
		t.Fatal("expected jobs from fixture")
	}
}

func TestSystemCrontabHasUserField(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "cron_d"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	var backup *model.Job
	for i := range result.Jobs {
		if result.Jobs[i].Command == "/opt/scripts/backup.sh" {
			backup = &result.Jobs[i]
			break
		}
	}
	if backup == nil {
		t.Fatal("backup job not found")
	}
	if backup.User != "root" {
		t.Fatalf("user = %q, want root", backup.User)
	}
	if backup.Command != "/opt/scripts/backup.sh" {
		t.Fatalf("command polluted with user field: %q", backup.Command)
	}
}

func TestCronDDotFileWarning(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "cron_d"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	for _, j := range result.Jobs {
		if j.Command == "/opt/scripts/never-runs.sh" {
			found := false
			for _, w := range j.Warnings {
				if w == "ignored_by_cron: files with '.' in name are skipped by cron" {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected ignored_by_cron warning, got %v", j.Warnings)
			}
			return
		}
	}
	t.Fatal("ignored.conf job not found")
}

func TestEnvLinesCaptured(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "user_crontab", User: "alice"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	var alice *model.Job
	for i := range result.Jobs {
		if result.Jobs[i].User == "alice" && result.Jobs[i].Command == "/home/alice/bin/daily-report.sh" {
			alice = &result.Jobs[i]
			break
		}
	}
	if alice == nil {
		t.Fatal("alice daily report not found")
	}
	if alice.Env["MAILTO"] != "admin@example.com" {
		t.Fatalf("MAILTO = %q", alice.Env["MAILTO"])
	}
	if alice.Comment != "Alice's daily report" {
		t.Fatalf("comment = %q", alice.Comment)
	}
}

func TestPercentTruncation(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "user_crontab", User: "alice"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	for _, j := range result.Jobs {
		if j.Schedule == "0 9 * * *" {
			if j.Command != "echo hello" {
				t.Fatalf("command = %q, want echo hello", j.Command)
			}
			return
		}
	}
	t.Fatal("percent job not found")
}

func TestRebootHasNoNextRun(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "cron_d"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	for _, j := range result.Jobs {
		if j.Schedule == "@reboot" {
			if j.NextRun != nil {
				t.Fatalf("expected nil NextRun for @reboot, got %v", j.NextRun)
			}
			return
		}
	}
	t.Fatal("@reboot job not found")
}

func TestJobIDStableAcrossScheduleChange(t *testing.T) {
	id1 := model.JobID("/etc/cron.d/backup", "root", "/opt/scripts/backup.sh")
	id2 := model.JobID("/etc/cron.d/backup", "root", "/opt/scripts/backup.sh")
	if id1 != id2 {
		t.Fatalf("IDs differ: %s vs %s", id1, id2)
	}
	// Schedule is not part of ID — same command/path/user always same ID.
	_ = model.JobID // schedule changes don't affect ID by design
}

func TestRHELCrontabPath(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "user_crontab", User: "bob"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(result.Jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(result.Jobs))
	}
	if result.Jobs[0].Command != "/home/bob/bin/sync.sh" {
		t.Fatalf("command = %q", result.Jobs[0].Command)
	}
}

func TestAnacronCollection(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "anacron"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(result.Jobs) < 3 {
		t.Fatalf("expected at least 3 anacron jobs, got %d", len(result.Jobs))
	}
}

func TestRunPartsCollection(t *testing.T) {
	root := fixtureRoot(t)
	result, err := Collect(Options{FSRoot: root, Source: "run_parts"})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(result.Jobs) < 2 {
		t.Fatalf("expected run-parts jobs, got %d", len(result.Jobs))
	}
}
