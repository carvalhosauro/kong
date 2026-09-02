package collect

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/carvalhosauro/kong/internal/model"
)

func collectEtcCrontab(fsRoot string) ([]model.Job, error) {
	path := joinFS(fsRoot, "etc", "crontab")
	return parseCrontabFile(path, formatSystem, model.SourceEtcCrontab, "root")
}

func collectCronD(fsRoot string) ([]model.Job, error) {
	dir := joinFS(fsRoot, "etc", "cron.d")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var jobs []model.Job
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(dir, name)
		fileJobs, err := parseCrontabFile(path, formatSystem, model.SourceCronD, "root")
		if err != nil {
			continue
		}
		jobs = append(jobs, fileJobs...)
	}
	return jobs, nil
}

func collectUserCrontabs(fsRoot string, filterUser string) ([]model.Job, []string, error) {
	var inaccessible []string
	var jobs []model.Job

	spoolPaths := []struct {
		path   string
		format string
	}{
		{joinFS(fsRoot, "var", "spool", "cron", "crontabs"), "debian"},
		{joinFS(fsRoot, "var", "spool", "cron"), "rhel"},
	}

	seen := make(map[string]bool)
	for _, spool := range spoolPaths {
		entries, err := os.ReadDir(spool.path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if os.IsPermission(err) {
				inaccessible = append(inaccessible, spool.path)
				continue
			}
			return nil, inaccessible, err
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			user := e.Name()
			if filterUser != "" && user != filterUser {
				continue
			}
			path := filepath.Join(spool.path, user)
			if seen[path] {
				continue
			}
			seen[path] = true

			fileJobs, err := parseCrontabFile(path, formatUser, model.SourceUserCrontab, user)
			if err != nil {
				if os.IsPermission(err) {
					inaccessible = append(inaccessible, path)
					continue
				}
				continue
			}
			jobs = append(jobs, fileJobs...)
		}
	}

	return jobs, inaccessible, nil
}
