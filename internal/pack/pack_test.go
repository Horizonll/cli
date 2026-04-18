package pack

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/Horizonll/cli/internal/sanitize"
)

func TestCreateWritesExpectedZipStructure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sessionID := "20260101T000000Z"
	sessionDir := filepath.Join(os.Getenv("HOME"), ".repro", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"command","command":"go test ./...","exit_code":1}` + "\n"
	if err := os.WriteFile(filepath.Join(sessionDir, "session.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out.zip")
	res, err := Create(Options{ID: sessionID, Out: out, Sanitize: sanitize.LevelBalanced, Minimal: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != out {
		t.Fatalf("unexpected output path: %s", res.Path)
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	required := map[string]bool{
		"repro-" + sessionID + "/repro.sh":             false,
		"repro-" + sessionID + "/session.jsonl":        false,
		"repro-" + sessionID + "/sanitize_report.json": false,
		"repro-" + sessionID + "/system.json":          false,
	}
	for _, f := range zr.File {
		if _, ok := required[f.Name]; ok {
			required[f.Name] = true
		}
	}
	for k, ok := range required {
		if !ok {
			t.Fatalf("missing zip entry: %s", k)
		}
	}
}
