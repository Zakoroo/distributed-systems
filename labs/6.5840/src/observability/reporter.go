// Package observability reports progress and events to a shared writer.
package observability

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

const barWidth = 24

// Reporter serializes progress updates and log lines on the same writer.
type Reporter struct {
	mu          sync.Mutex
	out         io.Writer
	logger      *log.Logger
	interactive bool
	last        string
	barVisible  bool
}

// New creates a reporter. Set interactive for terminals that support ANSI
// line clearing. Other writers receive one line per distinct progress update.
func New(out io.Writer, interactive bool) *Reporter {
	return &Reporter{out: out, logger: log.New(out, "", log.LstdFlags), interactive: interactive}
}

// Progress shows a snapshot of completed work. Completed and total must be
// nonnegative, with completed at most total. Zero total means no work remains.
func (r *Reporter) Progress(label string, completed, total int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	filled := barWidth
	if total > 0 {
		filled = int(float64(completed) / float64(total) * barWidth)
	}
	line := fmt.Sprintf("[%s%s] (total %d/%d) %s", strings.Repeat("#", filled), strings.Repeat("-", barWidth-filled), completed, total, label)
	if line == r.last {
		return
	}
	r.last = line
	r.clearBar()
	r.barVisible = r.interactive && completed < total
	if r.barVisible {
		fmt.Fprint(r.out, line)
	} else {
		fmt.Fprintln(r.out, line)
	}
}

// Logf writes a timestamped event and redraws the current bar when necessary.
func (r *Reporter) Logf(format string, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clearBar()
	r.logger.Printf(format, args...)
	if r.barVisible {
		fmt.Fprint(r.out, r.last)
	}
}

// clearBar clears the current terminal line when a progress bar is visible.
// The caller holds r.mu.
func (r *Reporter) clearBar() {
	if r.barVisible {
		fmt.Fprint(r.out, "\r\x1b[2K")
	}
}
