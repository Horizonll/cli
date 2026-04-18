package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Event struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp,omitempty"`
	Cwd       string `json:"cwd,omitempty"`
	Command   string `json:"command,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
	ID        string `json:"id,omitempty"`
	Shell     string `json:"shell,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
}

func LoadEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	out := make([]Event, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, fmt.Errorf("parse event: %w", err)
		}
		out = append(out, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
