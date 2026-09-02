package model

import "time"

type Source string

const (
	SourceUserCrontab  Source = "user_crontab"
	SourceEtcCrontab   Source = "etc_crontab"
	SourceCronD        Source = "cron_d"
	SourceRunParts     Source = "run_parts"
	SourceAnacron      Source = "anacron"
	SourceSystemdTimer Source = "systemd_timer"
)

type Status string

const (
	StatusUnknown Status = "unknown"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusRunning Status = "running"
	StatusMissed  Status = "missed"
)

type Confidence string

const (
	ConfidenceObserved Confidence = "observed"
	ConfidenceInferred Confidence = "inferred"
	ConfidenceUnknown  Confidence = "unknown"
)

type Job struct {
	ID            string            `json:"id"`
	Source        Source            `json:"source"`
	SourcePath    string            `json:"source_path"`
	User          string            `json:"user"`
	Schedule      string            `json:"schedule"`
	ScheduleHuman string            `json:"schedule_human"`
	Command       string            `json:"command"`
	Env           map[string]string `json:"env,omitempty"`
	Comment       string            `json:"comment,omitempty"`
	NextRun       *time.Time        `json:"next_run,omitempty"`
	LastStart     *time.Time        `json:"last_start,omitempty"`
	LastStatus    Status            `json:"last_status"`
	Confidence    Confidence        `json:"confidence"`
	Warnings      []string          `json:"warnings,omitempty"`
}

type CollectResult struct {
	Jobs                []Job    `json:"jobs"`
	Partial             bool     `json:"partial"`
	InaccessibleSources []string `json:"inaccessible_sources,omitempty"`
}
