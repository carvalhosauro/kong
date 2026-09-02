package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/carvalhosauro/kong/internal/model"
)

func JSON(w io.Writer, result model.CollectResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func Table(w io.Writer, result model.CollectResult) error {
	if result.Partial {
		_, _ = fmt.Fprintf(w, "partial: true (inaccessible: %s)\n", strings.Join(result.InaccessibleSources, ", "))
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "SOURCE\tUSER\tSCHEDULE\tNEXT\tLAST START\tSTATUS\tCOMMAND")
	for _, j := range result.Jobs {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			j.Source,
			j.User,
			j.Schedule,
			formatTime(j.NextRun),
			formatTime(j.LastStart),
			j.LastStatus,
			truncate(j.Command, 60),
		)
		for _, warn := range j.Warnings {
			_, _ = fmt.Fprintf(tw, "\t\t\t\t\twarning: %s\n", warn)
		}
	}
	return tw.Flush()
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(time.RFC3339)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
