package minimize

import (
	"testing"

	"github.com/Horizonll/cli/internal/model"
)

func TestApplyFiltersNoiseAndCollapsesCD(t *testing.T) {
	events := []model.Event{
		{Type: "command", Command: "ls"},
		{Type: "command", Command: "cd /tmp"},
		{Type: "command", Command: "cd /work"},
		{Type: "command", Command: "go test ./..."},
	}
	out := Apply(events, true)
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
	if out[0].Command != "cd /work" {
		t.Fatalf("expected collapsed cd, got %q", out[0].Command)
	}
}
