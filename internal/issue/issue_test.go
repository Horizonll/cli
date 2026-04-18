package issue

import (
	"strings"
	"testing"

	"github.com/Horizonll/cli/internal/model"
	"github.com/Horizonll/cli/internal/systeminfo"
)

func TestMarkdownContainsSteps(t *testing.T) {
	md := Markdown("id1", []model.Event{{Type: "command", Command: "go test ./...", ExitCode: 1}}, systeminfo.Info{OS: "linux", Arch: "amd64"})
	if !strings.Contains(md, "go test ./...") {
		t.Fatalf("expected command in markdown")
	}
}
