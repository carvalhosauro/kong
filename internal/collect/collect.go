package collect

import (
	"github.com/carvalhosauro/kong/internal/model"
)

type Options struct {
	FSRoot string
	User   string
	Source string
}

func Collect(opts Options) (model.CollectResult, error) {
	result := model.CollectResult{}
	var all []model.Job

	if shouldCollect(opts.Source, "etc_crontab") {
		jobs, err := collectEtcCrontab(opts.FSRoot)
		if err == nil {
			all = append(all, jobs...)
		}
	}

	if shouldCollect(opts.Source, "cron_d") {
		jobs, err := collectCronD(opts.FSRoot)
		if err == nil {
			all = append(all, jobs...)
		}
	}

	if shouldCollect(opts.Source, "user_crontab") {
		jobs, inaccessible, err := collectUserCrontabs(opts.FSRoot, opts.User)
		if err != nil {
			return result, err
		}
		all = append(all, jobs...)
		if len(inaccessible) > 0 {
			result.Partial = true
			result.InaccessibleSources = append(result.InaccessibleSources, inaccessible...)
		}
	}

	if shouldCollect(opts.Source, "run_parts") {
		jobs, err := collectRunParts(opts.FSRoot)
		if err == nil {
			all = append(all, jobs...)
		}
	}

	if shouldCollect(opts.Source, "anacron") {
		jobs, err := collectAnacron(opts.FSRoot)
		if err == nil {
			all = append(all, jobs...)
		}
	}

	if shouldCollect(opts.Source, "systemd_timer") && opts.FSRoot == "" {
		jobs, err := collectSystemdTimers()
		if err == nil {
			all = append(all, jobs...)
		}
	}

	if opts.FSRoot == "" {
		enrichLastStartFromLogs(all)
	}

	if opts.User != "" {
		filtered := all[:0]
		for _, j := range all {
			if j.User == opts.User {
				filtered = append(filtered, j)
			}
		}
		all = filtered
	}

	result.Jobs = all
	return result, nil
}

func shouldCollect(filter, name string) bool {
	if filter == "" {
		return true
	}
	return filter == name
}
