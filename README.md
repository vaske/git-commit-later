# git-commit-later

Schedule a staged Git commit for later.

```bash
git commit-later "feat: ship this later" --in 2h
git commit-later "docs: publish tomorrow" --at "2026-08-25 09:00"
```

`git-commit-later` captures the staged Git tree when you schedule the job. At execution time it creates a commit from that snapshot and advances the original branch only if the branch still points to the same commit it did when the job was scheduled.

## Install

Git exposes the binary as `git commit-later` once `git-commit-later` is on your `PATH`.

### Homebrew

```bash
brew tap vaske/git-commit-later https://github.com/vaske/git-commit-later.git
brew install --cask git-commit-later
```

### Install script

No Go required. Downloads the latest GitHub Release for your OS/architecture:

```bash
curl -fsSL https://raw.githubusercontent.com/vaske/git-commit-later/master/install.sh | sh
```

Installs to `/usr/local/bin` if writable, otherwise `~/.local/bin`. Override with `PREFIX=/path/to/bin`. Pin a version with `VERSION=v0.1.0`.

### From a release binary

Download the archive for your OS/architecture from [GitHub Releases](https://github.com/vaske/git-commit-later/releases), extract `git-commit-later`, and place it somewhere on your `PATH`.

### From source (Go)

```bash
go install github.com/milan/git-commit-later@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Usage

Stage exactly what you want to commit:

```bash
git add src/payment.go README.md
```

Schedule it:

```bash
git commit-later "feat: payment flow" --in 2h
```

Or use a local date/time:

```bash
git commit-later "feat: payment flow" --at "2026-08-25 09:00"
```

Inspect jobs:

```bash
git commit-later list
```

Cancel a pending job:

```bash
git commit-later cancel 78b13ca2
```

Run a due job manually:

```bash
git commit-later run 78b13ca2
```

## Safety model

A scheduled job stores:

- repository path
- branch
- HEAD at scheduling time
- staged Git tree
- commit message
- Git author identity
- execution time

If the branch moves before execution, the scheduled job fails instead of silently committing onto a different history.

This means the MVP favors safety over automatic rebasing.

## Current MVP limitation

The scheduler uses a detached local worker process. Closing your terminal is fine, but a machine reboot or shutdown before the scheduled time can prevent the job from running. Persistent OS schedulers (launchd/systemd/Task Scheduler) are a planned improvement.

## Development

```bash
go test ./...
go build ./...
```

To install your local checkout:

```bash
go install .
```

Then:

```bash
git commit-later --help
```

## Releasing

Push a version tag. GitHub Actions runs tests, then GoReleaser publishes binaries, checksums, and the Homebrew cask:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

MIT
