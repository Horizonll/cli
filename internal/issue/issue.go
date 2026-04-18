package issue

import (
	"fmt"
	"strings"

	"github.com/Horizonll/cli/internal/model"
	"github.com/Horizonll/cli/internal/systeminfo"
)

func Markdown(id string, events []model.Event, info systeminfo.Info) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Reproduction Session\n\n")
	fmt.Fprintf(&b, "- Session ID: `%s`\n", id)
	fmt.Fprintf(&b, "- OS/Arch: `%s/%s`\n", info.OS, info.Arch)
	fmt.Fprintf(&b, "- Shell: `%s`\n", info.Shell)
	fmt.Fprintf(&b, "- Host: `%s`\n", info.Hostname)
	fmt.Fprintf(&b, "\n## Steps\n")
	i := 1
	for _, ev := range events {
		if ev.Type != "command" {
			continue
		}
		fmt.Fprintf(&b, "%d. `%s` (exit: %d)\n", i, ev.Command, ev.ExitCode)
		i++
	}
	fmt.Fprintf(&b, "\n## Attachments\n- Attach the generated repro pack zip from `repro pack`.\n")
	return b.String()
}
