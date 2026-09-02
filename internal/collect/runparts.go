package collect

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/carvalhosauro/kong/internal/model"
	"github.com/carvalhosauro/kong/internal/schedule"
)

var runPartsSchedules = map[string]string{
	"hourly":  "@hourly",
	"daily":   "@daily",
	"weekly":  "@weekly",
	"monthly": "@monthly",
}

func collectRunParts(fsRoot string) ([]model.Job, error) {
	var jobs []model.Job
	for period, macro := range runPartsSchedules {
		dir := joinFS(fsRoot, "etc", "cron."+period)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			path := filepath.Join(dir, name)
			next, _ := schedule.NextRun(macro, time.Now(), time.Local)
			cmd := path
			jobs = append(jobs, model.Job{
				ID:            model.JobID(path, "root", cmd),
				Source:        model.SourceRunParts,
				SourcePath:    path,
				User:          "root",
				Schedule:      macro,
				ScheduleHuman: schedule.Humanize(macro),
				Command:       cmd,
				NextRun:       next,
				LastStatus:    model.StatusUnknown,
				Confidence:    model.ConfidenceUnknown,
			})
		}
	}
	return jobs, nil
}

func collectAnacron(fsRoot string) ([]model.Job, error) {
	path := joinFS(fsRoot, "etc", "anacrontab")
	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var jobs []model.Job
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// period delay job-id command...
		period := fields[0]
		if !isDigits(period) {
			continue
		}
		jobID := fields[2]
		command := strings.Join(fields[3:], " ")
		cmd := schedule.TruncateCommand(command)
		sched := "anacron:" + period + "d"
		jobs = append(jobs, model.Job{
			ID:            model.JobID(path, "root", cmd),
			Source:        model.SourceAnacron,
			SourcePath:    path,
			User:          "root",
			Schedule:      sched,
			ScheduleHuman: "every " + period + " days (anacron)",
			Command:       cmd,
			Comment:       jobID,
			LastStatus:    model.StatusUnknown,
			Confidence:    model.ConfidenceUnknown,
		})
	}
	return jobs, scanner.Err()
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
