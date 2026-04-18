# repro CLI (MVP)

`repro` captures terminal troubleshooting sessions and packages a sanitized, minimal, issue-ready repro bundle.

## Quickstart

```bash
go run ./cmd/repro shell
# reproduce your issue inside the managed shell, then exit

go run ./cmd/repro pack --sanitize balanced --minimal=true

go run ./cmd/repro issue
```

## Commands

- `repro shell` – PTY-managed subshell recording to `~/.repro/sessions/<id>/session.jsonl`
- `repro pack` – build zip from latest or specified session
  - `--out` output zip path
  - `--id` session id
  - `--sanitize strict|balanced|off` (default `balanced`)
  - `--minimal` (default `true`)
- `repro issue` – print GitHub issue-ready Markdown from latest or specified session

## Pack structure

```text
repro-<id>/
  README.md
  repro.sh
  steps.md
  issue.md
  system.json
  session.jsonl
  sanitize_report.json
  artifacts/stdout_tail.txt
```

## Privacy notes

Sanitization is enabled by default (`balanced`) and includes:

- GitHub tokens (`ghp_`, `github_pat_`)
- Bearer authorization headers
- AWS access keys (`AKIA`/`ASIA`)
- Private key blocks
- Common `token/password/apikey/secret` key-value patterns

A `sanitize_report.json` file is emitted for auditability. Always review generated output before sharing.

## Development and release

See `CONTRIBUTING.md` for:

- how to verify active PR/task status and confirm base branch is `main`
- local validation commands (`go test`, `go install`, `repro shell/pack/issue`)
- release tagging and GoReleaser workflow verification steps
