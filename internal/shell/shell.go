package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Horizonll/cli/internal/model"
	"github.com/Horizonll/cli/internal/session"
	"github.com/creack/pty"
)

func Run() (string, error) {
	if _, err := session.EnsureRoot(); err != nil {
		return "", err
	}
	id := session.NewID()
	dir, err := session.SessionDir(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	jsonlPath := filepath.Join(dir, "session.jsonl")
	f, err := os.OpenFile(jsonlPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/bash"
	}
	base := filepath.Base(shellPath)

	if err := writeEvent(f, model.Event{Type: "meta", Timestamp: time.Now().UTC().Format(time.RFC3339), ID: id, Shell: shellPath}); err != nil {
		return "", err
	}

	initPath, zdotdir, err := writeInitFiles(dir, jsonlPath)
	if err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	if strings.Contains(base, "zsh") {
		cmd = exec.Command(shellPath, "-i")
		cmd.Env = append(os.Environ(), "REPRO_SESSION_FILE="+jsonlPath, "ZDOTDIR="+zdotdir)
	} else {
		cmd = exec.Command(shellPath, "--rcfile", initPath, "-i")
		cmd.Env = append(os.Environ(), "REPRO_SESSION_FILE="+jsonlPath)
	}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return "", err
	}
	defer func() { _ = ptmx.Close() }()

	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}
	if writeErr := writeEvent(f, model.Event{Type: "end", Timestamp: time.Now().UTC().Format(time.RFC3339), EndedAt: time.Now().UTC().Format(time.RFC3339), ExitCode: exitCode}); writeErr != nil {
		return id, writeErr
	}
	return id, nil
}

func writeEvent(w io.Writer, ev model.Event) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", string(b))
	return err
}

func writeInitFiles(sessionDir, jsonlPath string) (string, string, error) {
	initContent := `repro_escape() {
local s="$1"
s=${s//\\/\\\\}
s=${s//\"/\\\"}
s=${s//$'\n'/\\n}
s=${s//$'\r'/\\r}
printf '%s' "$s"
}

repro_log() {
local ec="$1"
local cmd="$2"
local cwd="$PWD"
local ts
ts=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
printf '{"type":"command","timestamp":"%s","cwd":"%s","command":"%s","exit_code":%s}\n' \
"$ts" "$(repro_escape "$cwd")" "$(repro_escape "$cmd")" "$ec" >> "$REPRO_SESSION_FILE"
}

if [ -n "$BASH_VERSION" ]; then
__repro_last_cmd=""
	__repro_preexec() {
		case "$BASH_COMMAND" in
			repro_*|__repro_*|history*|PROMPT_COMMAND=*|'[ -n "$ZSH_VERSION" ]'|'[ -n "$BASH_VERSION" ]') return ;;
		esac
		__repro_last_cmd="$BASH_COMMAND"
	}
__repro_precmd() {
local ec=$?
if [ -n "$__repro_last_cmd" ]; then
repro_log "$ec" "$__repro_last_cmd"
__repro_last_cmd=""
fi
}
trap '__repro_preexec' DEBUG
PROMPT_COMMAND='__repro_precmd'
fi

if [ -n "$ZSH_VERSION" ]; then
autoload -Uz add-zsh-hook
__repro_last_cmd=""
__repro_preexec() { __repro_last_cmd="$1"; }
__repro_precmd() {
local ec=$?
if [ -n "$__repro_last_cmd" ]; then
repro_log "$ec" "$__repro_last_cmd"
__repro_last_cmd=""
fi
}
add-zsh-hook preexec __repro_preexec
add-zsh-hook precmd __repro_precmd
fi
`
	initPath := filepath.Join(sessionDir, ".repro_init.sh")
	if err := os.WriteFile(initPath, []byte(initContent), 0o600); err != nil {
		return "", "", err
	}
	zdotdir := filepath.Join(sessionDir, ".zdot")
	if err := os.MkdirAll(zdotdir, 0o700); err != nil {
		return "", "", err
	}
	zshrc := fmt.Sprintf("export REPRO_SESSION_FILE=%q\nsource %q\n", jsonlPath, initPath)
	if err := os.WriteFile(filepath.Join(zdotdir, ".zshrc"), []byte(zshrc), 0o600); err != nil {
		return "", "", err
	}
	return initPath, zdotdir, nil
}
