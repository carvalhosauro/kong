package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var macroSchedules = map[string]string{
	"@reboot":   "",
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

func IsReboot(schedule string) bool {
	return strings.EqualFold(strings.TrimSpace(schedule), "@reboot")
}

func Normalize(schedule string) (string, bool, error) {
	s := strings.TrimSpace(schedule)
	if s == "" {
		return "", false, fmt.Errorf("empty schedule")
	}
	if strings.HasPrefix(s, "@") {
		if mapped, ok := macroSchedules[strings.ToLower(s)]; ok {
			if mapped == "" {
				return s, true, nil
			}
			return mapped, false, nil
		}
		return "", false, fmt.Errorf("unknown macro: %s", s)
	}
	fields := strings.Fields(s)
	if len(fields) != 5 {
		return "", false, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}
	// Vixie cron allows 7 for Sunday; robfig uses 0.
	if fields[4] == "7" {
		fields[4] = "0"
		s = strings.Join(fields, " ")
	}
	return s, false, nil
}

func NextRun(schedule string, from time.Time, loc *time.Location) (*time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	normalized, reboot, err := Normalize(schedule)
	if err != nil {
		return nil, err
	}
	if reboot {
		return nil, nil
	}

	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	sched, err := parser.Parse(normalized)
	if err != nil {
		return nil, err
	}

	next := sched.Next(from.In(loc))
	if next.IsZero() {
		return nil, nil
	}
	t := next.In(loc)
	return &t, nil
}

func Humanize(schedule string) string {
	s := strings.ToLower(strings.TrimSpace(schedule))
	switch s {
	case "@reboot":
		return "at reboot"
	case "@yearly", "@annually":
		return "yearly on Jan 1 at 00:00"
	case "@monthly":
		return "monthly on day 1 at 00:00"
	case "@weekly":
		return "weekly on Sunday at 00:00"
	case "@daily", "@midnight":
		return "daily at 00:00"
	case "@hourly":
		return "hourly"
	default:
		return schedule
	}
}

func TruncateCommand(command string) string {
	var b strings.Builder
	escaped := false
	for _, r := range command {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			b.WriteRune(r)
			continue
		}
		if r == '%' {
			break
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
