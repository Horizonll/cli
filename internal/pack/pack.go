package pack

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Horizonll/cli/internal/issue"
	"github.com/Horizonll/cli/internal/minimize"
	"github.com/Horizonll/cli/internal/model"
	"github.com/Horizonll/cli/internal/sanitize"
	"github.com/Horizonll/cli/internal/session"
	"github.com/Horizonll/cli/internal/systeminfo"
)

type Options struct {
	ID       string
	Out      string
	Sanitize sanitize.Level
	Minimal  bool
}

type Result struct {
	Path string
	ID   string
}

func Create(opts Options) (Result, error) {
	id, err := session.ResolveID(opts.ID)
	if err != nil {
		return Result{}, err
	}
	jsonlPath, err := session.SessionJSONLPath(id)
	if err != nil {
		return Result{}, err
	}
	events, err := model.LoadEvents(jsonlPath)
	if err != nil {
		return Result{}, err
	}

	sz := sanitize.New(opts.Sanitize)
	sanitizedEvents := make([]model.Event, 0, len(events))
	reports := make([]sanitize.Report, 0, len(events))
	for _, ev := range events {
		if ev.Type == "command" {
			cmd, rep := sz.Apply(ev.Command)
			ev.Command = cmd
			reports = append(reports, rep)
		}
		sanitizedEvents = append(sanitizedEvents, ev)
	}
	sanitizedEvents = minimize.Apply(sanitizedEvents, opts.Minimal)
	report := sanitize.MergeReports(reports...)

	info := systeminfo.Collect()
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	sessionJSONL, err := marshalJSONL(sanitizedEvents)
	if err != nil {
		return Result{}, err
	}
	script := buildScript(sanitizedEvents)
	steps := buildSteps(sanitizedEvents)
	issueMD := issue.Markdown(id, sanitizedEvents, info)
	readme := buildReadme(id)

	out := opts.Out
	if strings.TrimSpace(out) == "" {
		out = fmt.Sprintf("repro-%s.zip", id)
	}
	if err := writeZip(out, id, map[string][]byte{
		"README.md":                 []byte(readme),
		"repro.sh":                  []byte(script),
		"steps.md":                  []byte(steps),
		"issue.md":                  []byte(issueMD),
		"system.json":               infoJSON,
		"session.jsonl":             sessionJSONL,
		"sanitize_report.json":      reportJSON,
		"artifacts/stdout_tail.txt": []byte("See session.jsonl command sequence for MVP output evidence.\n"),
	}); err != nil {
		return Result{}, err
	}
	return Result{Path: out, ID: id}, nil
}

func marshalJSONL(events []model.Event) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func buildScript(events []model.Event) string {
	lines := []string{"#!/usr/bin/env bash", "set -euo pipefail", ""}
	for _, ev := range events {
		if ev.Type != "command" {
			continue
		}
		lines = append(lines, ev.Command)
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildSteps(events []model.Event) string {
	lines := []string{"# Repro Steps", ""}
	n := 1
	for _, ev := range events {
		if ev.Type != "command" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. `%s` (exit: %d)", n, ev.Command, ev.ExitCode))
		n++
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildReadme(id string) string {
	return fmt.Sprintf("# Repro Pack %s\n\nRun `bash repro.sh` to replay the sanitized minimal session.\n", id)
}

func writeZip(path string, id string, files map[string][]byte) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	root := fmt.Sprintf("repro-%s", id)
	for name, body := range files {
		w, err := zw.Create(filepath.ToSlash(filepath.Join(root, name)))
		if err != nil {
			_ = zw.Close()
			return err
		}
		if _, err := w.Write(body); err != nil {
			_ = zw.Close()
			return err
		}
	}
	return zw.Close()
}
