# NEXT_STEPS

## Local verification

Run from repository root:

```bash
go test ./...
go vet ./...
```

Quick CLI flow:

```bash
go run ./cmd/repro --help
go run ./cmd/repro shell
go run ./cmd/repro pack --out ./repro-last.zip
unzip -l ./repro-last.zip
go run ./cmd/repro issue
```

## Release (tag `v*`)

Release automation is triggered by pushing a version tag that matches `v*`.

```bash
git checkout main
git pull --ff-only
git tag v0.1.0
git push origin v0.1.0
```

Then verify artifacts in GitHub Releases (darwin/linux, amd64/arm64, checksums).
