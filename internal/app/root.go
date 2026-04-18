package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Horizonll/cli/internal/issue"
	"github.com/Horizonll/cli/internal/minimize"
	"github.com/Horizonll/cli/internal/model"
	"github.com/Horizonll/cli/internal/pack"
	"github.com/Horizonll/cli/internal/sanitize"
	"github.com/Horizonll/cli/internal/session"
	"github.com/Horizonll/cli/internal/shell"
	"github.com/Horizonll/cli/internal/systeminfo"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "repro",
		Short: "Capture and share reproducible terminal sessions",
	}
	root.AddCommand(newShellCmd())
	root.AddCommand(newPackCmd())
	root.AddCommand(newIssueCmd())
	return root
}

func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Start a PTY-managed shell and record commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := shell.Run()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Recorded session: %s\n", id)
			return nil
		},
	}
}

func newPackCmd() *cobra.Command {
	var (
		id      string
		out     string
		sLevel  string
		minimal bool
	)
	c := &cobra.Command{
		Use:   "pack",
		Short: "Create a repro zip pack for the latest or specified session",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := pack.Create(pack.Options{
				ID:       id,
				Out:      out,
				Sanitize: sanitize.ParseLevel(sLevel),
				Minimal:  minimal,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created pack: %s\n", res.Path)
			return nil
		},
	}
	c.Flags().StringVar(&id, "id", "", "Session ID (default latest)")
	c.Flags().StringVar(&out, "out", "", "Output zip path")
	c.Flags().StringVar(&sLevel, "sanitize", string(sanitize.LevelBalanced), "Sanitize level: strict|balanced|off")
	c.Flags().BoolVar(&minimal, "minimal", true, "Enable command minimalization")
	return c
}

func newIssueCmd() *cobra.Command {
	var (
		id      string
		minimal bool
	)
	c := &cobra.Command{
		Use:   "issue",
		Short: "Render GitHub issue-ready Markdown for a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := session.ResolveID(id)
			if err != nil {
				return err
			}
			path, err := session.SessionJSONLPath(resolved)
			if err != nil {
				return err
			}
			events, err := model.LoadEvents(path)
			if err != nil {
				return err
			}
			sz := sanitize.New(sanitize.LevelBalanced)
			for i := range events {
				if events[i].Type == "command" {
					events[i].Command, _ = sz.Apply(events[i].Command)
				}
			}
			events = minimize.Apply(events, minimal)
			md := issue.Markdown(resolved, events, systeminfo.Collect())
			_, _ = cmd.OutOrStdout().Write([]byte(md))
			return nil
		},
	}
	c.Flags().StringVar(&id, "id", "", "Session ID (default latest)")
	c.Flags().BoolVar(&minimal, "minimal", true, "Enable command minimalization")
	return c
}

func WriteSystemJSON(path string) error {
	info := systeminfo.Collect()
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
