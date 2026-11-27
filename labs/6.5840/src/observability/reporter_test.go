package observability

import (
	"bytes"
	"strings"
	"testing"
)

func TestRedirectedProgressUsesDistinctLines(t *testing.T) {
	var out bytes.Buffer
	r := New(&out, false)
	r.Progress("map", 1, 2)
	r.Progress("map", 1, 2)
	r.Logf("retrying task %d", 0)
	r.Progress("map", 2, 2)

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 3 || strings.ContainsAny(out.String(), "\r\x1b") {
		t.Fatalf("expected distinct plain progress and log lines, got %q", out.String())
	}
	if !strings.HasSuffix(lines[0], "(total 1/2) map") || !strings.HasSuffix(lines[1], "retrying task 0") || !strings.HasSuffix(lines[2], "(total 2/2) map") {
		t.Fatalf("unexpected progress or events: %q", lines)
	}
}

func TestTerminalLogRedrawsActiveBar(t *testing.T) {
	var out bytes.Buffer
	r := New(&out, true)
	r.Progress("work", 1, 2)
	bar := out.String()
	r.Logf("retrying")
	if !strings.Contains(out.String(), bar+"\r\x1b[2K") || !strings.HasSuffix(out.String(), "retrying\n"+bar) {
		t.Fatalf("expected the log line to clear and redraw the bar, got %q", out.String())
	}

	r.Progress("work", 2, 2)
	r.Logf("complete")
	if !strings.HasSuffix(out.String(), "complete\n") {
		t.Fatalf("expected the finished bar to stay on its own line, got %q", out.String())
	}
}
