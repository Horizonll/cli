package minimize

import (
	"strings"

	"github.com/Horizonll/cli/internal/model"
)

var noisy = map[string]struct{}{
	"ls": {}, "pwd": {}, "clear": {}, "whoami": {}, "date": {}, "history": {},
}

func Apply(events []model.Event, enabled bool) []model.Event {
	if !enabled {
		return events
	}
	out := make([]model.Event, 0, len(events))
	prevWasCD := false
	for _, ev := range events {
		if ev.Type != "command" {
			out = append(out, ev)
			continue
		}
		trimmed := strings.TrimSpace(ev.Command)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		cmd := fields[0]
		if _, ok := noisy[cmd]; ok {
			continue
		}
		if cmd == "cd" {
			if prevWasCD {
				if len(out) > 0 {
					out[len(out)-1] = ev
				}
				continue
			}
			prevWasCD = true
		} else {
			prevWasCD = false
		}
		out = append(out, ev)
	}
	return out
}
