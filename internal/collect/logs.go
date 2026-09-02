package collect

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/carvalhosauro/kong/internal/model"
)

var cronLogPattern = regexp.MustCompile(`\(([^)]+)\)\s+CMD\s+\((.+)\)`)

// enrichLastStartFromLogs best-effort matches journal/syslog CRON lines to jobs.
func enrichLastStartFromLogs(jobs []model.Job) {
	entries := readCronLogEntries()
	if len(entries) == 0 {
		return
	}

	for i := range jobs {
		if jobs[i].Source == model.SourceSystemdTimer {
			continue
		}
		cmd := strings.TrimSpace(jobs[i].Command)
		user := jobs[i].User
		var best *time.Time
		for _, e := range entries {
			if user != "" && e.user != user {
				continue
			}
			if !commandMatches(e.command, cmd) {
				continue
			}
			if best == nil || e.at.After(*best) {
				t := e.at
				best = &t
			}
		}
		if best != nil {
			jobs[i].LastStart = best
			jobs[i].Confidence = model.ConfidenceInferred
		}
	}
}

type cronLogEntry struct {
	at      time.Time
	user    string
	command string
}

func readCronLogEntries() []cronLogEntry {
	if entries := readJournalCronEntries(); len(entries) > 0 {
		return entries
	}
	return readSyslogCronEntries()
}

func readJournalCronEntries() []cronLogEntry {
	if _, err := exec.LookPath("journalctl"); err != nil {
		return nil
	}
	out, err := exec.Command("journalctl", "-t", "CRON", "--no-pager", "-o", "short-iso", "-n", "500").Output()
	if err != nil {
		return nil
	}
	return parseCronLogOutput(string(out))
}

func readSyslogCronEntries() []cronLogEntry {
	paths := []string{"/var/log/syslog", "/var/log/cron", "/var/log/messages"}
	for _, path := range paths {
		if entries := readSyslogFile(path); entries != nil {
			return entries
		}
	}
	return nil
}

func readSyslogFile(path string) []cronLogEntry {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 500 {
		lines = lines[len(lines)-500:]
	}
	return parseCronLogOutput(strings.Join(lines, "\n"))
}

func parseCronLogOutput(content string) []cronLogEntry {
	var entries []cronLogEntry
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "CRON") && !strings.Contains(line, "cron") {
			continue
		}
		m := cronLogPattern.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		at := parseLogTimestamp(line)
		if at.IsZero() {
			continue
		}
		entries = append(entries, cronLogEntry{
			at:      at,
			user:    m[1],
			command: m[2],
		})
	}
	return entries
}

func parseLogTimestamp(line string) time.Time {
	// ISO from journalctl short-iso: 2026-01-01T12:00:00+0000 host ...
	if len(line) >= 25 {
		candidate := line[:25]
		if t, err := time.Parse("2006-01-02T15:04:05-0700", candidate); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, strings.Fields(line)[0]); err == nil {
			return t
		}
	}
	return time.Time{}
}

func commandMatches(logCmd, jobCmd string) bool {
	logCmd = strings.TrimSpace(logCmd)
	jobCmd = strings.TrimSpace(jobCmd)
	if logCmd == jobCmd {
		return true
	}
	return strings.Contains(logCmd, jobCmd) || strings.Contains(jobCmd, logCmd)
}
