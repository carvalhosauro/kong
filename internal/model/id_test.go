package model

import "testing"

func TestJobIDExcludesSchedule(t *testing.T) {
	base := JobID("/etc/cron.d/job", "root", "/opt/run.sh")
	// Same path/user/command must produce same ID regardless of how we think about schedule.
	again := JobID("/etc/cron.d/job", "root", "/opt/run.sh")
	if base != again {
		t.Fatalf("unstable ID: %s vs %s", base, again)
	}

	withExtraSpace := JobID("/etc/cron.d/job", "root", "  /opt/run.sh  ")
	if base != withExtraSpace {
		t.Fatalf("normalization failed: %s vs %s", base, withExtraSpace)
	}

	withStdin := JobID("/etc/cron.d/job", "root", "/opt/run.sh%ignored stdin")
	if base != withStdin {
		t.Fatalf("%% truncation failed: %s vs %s", base, withStdin)
	}
}
