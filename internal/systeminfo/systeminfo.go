package systeminfo

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Info struct {
	OS       string            `json:"os"`
	Arch     string            `json:"arch"`
	Shell    string            `json:"shell"`
	Hostname string            `json:"hostname"`
	User     string            `json:"user"`
	Tools    map[string]string `json:"tools"`
}

func Collect() Info {
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	info := Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Shell:    os.Getenv("SHELL"),
		Hostname: hostname,
		User:     user,
		Tools:    map[string]string{},
	}
	for _, t := range []struct {
		name string
		args []string
	}{
		{name: "git", args: []string{"--version"}},
		{name: "go", args: []string{"version"}},
		{name: "node", args: []string{"--version"}},
		{name: "python", args: []string{"--version"}},
		{name: "docker", args: []string{"--version"}},
	} {
		if v, ok := version(t.name, t.args...); ok {
			info.Tools[t.name] = v
		}
	}
	return info
}

func version(bin string, args ...string) (string, bool) {
	cmd := exec.Command(bin, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", false
	}
	return strings.TrimSpace(out.String()), true
}
