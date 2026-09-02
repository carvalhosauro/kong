package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/carvalhosauro/kong/internal/collect"
	"github.com/carvalhosauro/kong/internal/render"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("kong %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	case "list":
		if err := runList(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "kong: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "kong: unknown command %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func runList(args []string) error {
	var jsonOut bool
	var user, source, fsRoot string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--user":
			i++
			if i >= len(args) {
				return fmt.Errorf("--user requires a value")
			}
			user = args[i]
		case "--source":
			i++
			if i >= len(args) {
				return fmt.Errorf("--source requires a value")
			}
			source = args[i]
		case "--fs-root":
			i++
			if i >= len(args) {
				return fmt.Errorf("--fs-root requires a value")
			}
			fsRoot = args[i]
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
	}

	result, err := collect.Collect(collect.Options{
		FSRoot: fsRoot,
		User:   user,
		Source: source,
	})
	if err != nil {
		return err
	}

	if jsonOut {
		return render.JSON(os.Stdout, result)
	}
	return render.Table(os.Stdout, result)
}

func usage() {
	fmt.Fprintf(os.Stderr, `kong — cron job inventory for Linux

Usage:
  kong list [flags]

Flags:
  --json          Output JSON
  --user USER     Filter by user
  --source SOURCE Filter by source (user_crontab, etc_crontab, cron_d, run_parts, anacron, systemd_timer)
  --fs-root PATH  Filesystem root for fixtures/testing

`)
}
