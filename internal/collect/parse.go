package collect

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/carvalhosauro/kong/internal/model"
	"github.com/carvalhosauro/kong/internal/schedule"
)

type crontabFormat int

const (
	formatUser crontabFormat = iota
	formatSystem
)

type parsedLine struct {
	schedule string
	user     string
	command  string
}

type fileEnv struct {
	vars map[string]string
	tz   *time.Location
}

func parseCrontabFile(path string, format crontabFormat, source model.Source, defaultUser string) ([]model.Job, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return parseCrontabContent(string(data), path, format, source, defaultUser)
}

func parseCrontabContent(content, path string, format crontabFormat, source model.Source, defaultUser string) ([]model.Job, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	env := fileEnv{vars: make(map[string]string), tz: time.Local}

	var jobs []model.Job
	var pendingComment string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			pendingComment = ""
			continue
		}
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if pendingComment != "" {
				pendingComment += " " + comment
			} else {
				pendingComment = comment
			}
			continue
		}

		if kv, ok := parseEnvLine(line); ok {
			env.vars[kv.key] = kv.value
			if kv.key == "CRON_TZ" {
				loc, err := time.LoadLocation(kv.value)
				if err == nil {
					env.tz = loc
				}
			}
			pendingComment = ""
			continue
		}

		parsed, err := parseJobLine(line, format)
		if err != nil {
			continue
		}

		user := parsed.user
		if user == "" {
			user = defaultUser
		}

		cmd := schedule.TruncateCommand(parsed.command)
		warnings := cronDWarnings(source, path, parsed.schedule)

		next, _ := schedule.NextRun(parsed.schedule, time.Now(), env.tz)
		human := schedule.Humanize(parsed.schedule)

		job := model.Job{
			ID:            model.JobID(path, user, cmd),
			Source:        source,
			SourcePath:    path,
			User:          user,
			Schedule:      parsed.schedule,
			ScheduleHuman: human,
			Command:       cmd,
			Env:           copyEnv(env.vars),
			Comment:       pendingComment,
			NextRun:       next,
			LastStatus:    model.StatusUnknown,
			Confidence:    model.ConfidenceUnknown,
			Warnings:      warnings,
		}
		jobs = append(jobs, job)
		pendingComment = ""
	}

	return jobs, scanner.Err()
}

type envKV struct {
	key, value string
}

func parseEnvLine(line string) (envKV, bool) {
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return envKV{}, false
	}
	key := strings.TrimSpace(line[:idx])
	if !isEnvKey(key) {
		return envKV{}, false
	}
	return envKV{key: key, value: strings.TrimSpace(line[idx+1:])}, true
}

func isEnvKey(key string) bool {
	for _, r := range key {
		if (r >= 'A' && r <= 'Z') || r == '_' {
			continue
		}
		return false
	}
	return len(key) > 0
}

func parseJobLine(line string, format crontabFormat) (parsedLine, error) {
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return parsedLine{}, fmt.Errorf("empty line")
	}

	// Macro or 5-field schedule.
	if strings.HasPrefix(fields[0], "@") {
		return parseMacroLine(fields, format)
	}
	if len(fields) < 6 && format == formatSystem {
		return parsedLine{}, fmt.Errorf("system crontab needs user field")
	}
	if len(fields) < 5 {
		return parsedLine{}, fmt.Errorf("not enough fields")
	}

	sched := strings.Join(fields[0:5], " ")
	switch format {
	case formatSystem:
		if len(fields) < 7 {
			return parsedLine{}, fmt.Errorf("system crontab needs user and command")
		}
		return parsedLine{
			schedule: sched,
			user:     fields[5],
			command:  strings.Join(fields[6:], " "),
		}, nil
	default:
		return parsedLine{
			schedule: sched,
			command:  strings.Join(fields[5:], " "),
		}, nil
	}
}

func parseMacroLine(fields []string, format crontabFormat) (parsedLine, error) {
	macro := fields[0]
	switch format {
	case formatSystem:
		if len(fields) < 3 {
			return parsedLine{}, fmt.Errorf("system macro line too short")
		}
		return parsedLine{
			schedule: macro,
			user:     fields[1],
			command:  strings.Join(fields[2:], " "),
		}, nil
	default:
		if len(fields) < 2 {
			return parsedLine{}, fmt.Errorf("user macro line too short")
		}
		return parsedLine{
			schedule: macro,
			command:  strings.Join(fields[1:], " "),
		}, nil
	}
}

func cronDWarnings(source model.Source, path string, sched string) []string {
	if source != model.SourceCronD {
		return nil
	}
	base := filepathBase(path)
	if isDotFile(base) {
		return []string{"ignored_by_cron: files with '.' in name are skipped by cron"}
	}
	_ = sched
	return nil
}

func filepathBase(path string) string {
	return filepath.Base(path)
}

func copyEnv(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
