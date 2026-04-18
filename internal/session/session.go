package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const dirPerm = 0o700

func rootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".repro", "sessions"), nil
}

func EnsureRoot() (string, error) {
	root, err := rootDir()
	if err != nil {
		return "", err
	}
	return root, os.MkdirAll(root, dirPerm)
}

func NewID() string {
	return time.Now().UTC().Format("20060102T150405Z")
}

func SessionDir(id string) (string, error) {
	root, err := rootDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(id) == "" {
		return "", errors.New("empty session id")
	}
	return filepath.Join(root, id), nil
}

func LatestID() (string, error) {
	root, err := rootDir()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("no sessions found")
		}
		return "", err
	}
	type candidate struct {
		id  string
		mod time.Time
	}
	items := make([]candidate, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, candidate{id: e.Name(), mod: info.ModTime()})
	}
	if len(items) == 0 {
		return "", errors.New("no sessions found")
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.After(items[j].mod) })
	return items[0].id, nil
}

func SessionJSONLPath(id string) (string, error) {
	dir, err := SessionDir(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.jsonl"), nil
}

func ResolveID(specified string) (string, error) {
	if strings.TrimSpace(specified) != "" {
		return specified, nil
	}
	id, err := LatestID()
	if err != nil {
		return "", fmt.Errorf("resolve session id: %w", err)
	}
	return id, nil
}
