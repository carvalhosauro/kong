package collect

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/carvalhosauro/kong/internal/model"
)

func collectSystemdTimers() ([]model.Job, error) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return nil, nil
	}

	out, err := exec.Command("systemctl", "list-timers", "--all", "--no-pager", "--no-legend").Output()
	if err != nil {
		return nil, nil
	}

	var jobs []model.Job
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		timerUnit := fields[0]
		if !strings.HasSuffix(timerUnit, ".timer") {
			continue
		}

		job, ok := systemdTimerJob(timerUnit)
		if ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, scanner.Err()
}

func systemdTimerJob(timerUnit string) (model.Job, bool) {
	props := querySystemdProps(timerUnit,
		"NextElapseUSecRealtime",
		"LastTriggerUSec",
		"Unit",
	)

	serviceUnit := strings.TrimSpace(props["Unit"])
	if serviceUnit == "" {
		serviceUnit = strings.TrimSuffix(timerUnit, ".timer") + ".service"
	}

	svcProps := querySystemdProps(serviceUnit,
		"ExecMainStatus",
		"Result",
		"ExecMainExitTimestamp",
	)

	var nextRun, lastStart *time.Time
	if t := parseSystemdTime(props["NextElapseUSecRealtime"]); t != nil {
		nextRun = t
	}
	if t := parseSystemdTime(props["LastTriggerUSec"]); t != nil {
		lastStart = t
	} else if t := parseSystemdTime(svcProps["ExecMainExitTimestamp"]); t != nil {
		lastStart = t
	}

	status := model.StatusUnknown
	confidence := model.ConfidenceObserved
	if result := svcProps["Result"]; result != "" {
		switch result {
		case "success":
			status = model.StatusSuccess
		case "failed", "exit-code":
			status = model.StatusFailed
		default:
			if result != "inactive" && result != "done" {
				status = model.StatusUnknown
			}
		}
	}

	cmd := "systemd:" + serviceUnit
	path := "systemd:" + timerUnit
	return model.Job{
		ID:            model.JobID(path, "", cmd),
		Source:        model.SourceSystemdTimer,
		SourcePath:    path,
		User:          "system",
		Schedule:      timerUnit,
		ScheduleHuman: "systemd timer",
		Command:       cmd,
		NextRun:       nextRun,
		LastStart:     lastStart,
		LastStatus:    status,
		Confidence:    confidence,
	}, true
}

func querySystemdProps(unit string, props ...string) map[string]string {
	args := []string{"show", unit, "--no-pager"}
	for _, p := range props {
		args = append(args, "--property="+p)
	}
	out, err := exec.Command("systemctl", args...).Output()
	if err != nil {
		return nil
	}
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func parseSystemdTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" || value == "n/a" {
		return nil
	}
	if strings.Contains(value, "-") {
		layouts := []string{
			time.RFC3339Nano,
			"Mon 2006-01-02 15:04:05 MST",
			"Mon 2006-01-02 15:04:05 UTC",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, value); err == nil {
				return &t
			}
		}
		return nil
	}
	usec, err := strconv.ParseInt(value, 10, 64)
	if err != nil || usec <= 0 {
		return nil
	}
	t := time.Unix(0, usec*1000)
	return &t
}
