# Contributing

## What to do next (for the current MVP PRs)

1. Check current open PRs and make sure implementation work is not duplicated.
2. If an existing coding-agent PR is still queued/in-progress, add notes/docs only (this file + README updates), then wait for that PR to finish.
3. Ensure every feature PR targets `main` as the base branch.

## Verify PR status and base branch

- GitHub UI: open the PR page and confirm **base: `main`**.
- Optional with GitHub CLI:

```bash
gh pr list --repo Horizonll/cli
gh pr view <PR_NUMBER> --repo Horizonll/cli --json baseRefName,headRefName,state,url
```

Expected: `baseRefName` is `main`.

## Local validation checklist

From repo root:

```bash
go test ./...
go install ./cmd/repro
```

Run the CLI end-to-end:

```bash
# start managed shell recording
repro shell
# run a few commands in the subshell, then exit

# pack the latest session
repro pack --out ./repro-last.zip --sanitize balanced --minimal=true
unzip -l ./repro-last.zip

# print issue markdown
repro issue
```

## Release process

1. Merge the PR into `main` after CI is green.
2. Create and push a semantic version tag:

```bash
git checkout main
git pull --ff-only
git tag v0.1.0
git push origin v0.1.0
```

3. Verify the `Release` workflow succeeded:
   - GitHub Actions → `Release` workflow run for the tag.
   - Confirm artifacts were published to GitHub Releases.
   - Confirm checksums file exists.

## CI workflows in this repository

- `CI` (`.github/workflows/ci.yml`): runs tests + golangci-lint on PRs and pushes to `main`.
- `Release` (`.github/workflows/release.yml`): runs GoReleaser on pushed tags matching `v*`.
