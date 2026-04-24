# px

`px` is a small Go CLI for switching terminal proxy environment variables on and off.

It detects common local proxy tools, writes a local config, and installs a shell function so `px on` and `px off` affect the current terminal session.

## Install

Download a binary for your platform from [GitHub Releases](https://github.com/hczs/px/releases).

Developers can also install from source:

```bash
go install github.com/hczs/px@latest
```

## Quick Start

Initialize `px`:

```bash
px init
```

Restart your terminal, or source the shell profile shown by `px init`.

Enable proxy:

```bash
px on
```

Check status:

```bash
px status
```

Test proxy:

```bash
px test
```

Disable proxy:

```bash
px off
```

## Commands

```text
px init      Detect shell/proxy, save config, install shell function
px on        Enable proxy in the current shell through the installed function
px off       Restore or clear proxy variables in the current shell
px status    Show current proxy environment
px test      Test current proxy with https://ipinfo.io/json
px list      List detected local proxy software
px config    Reconfigure proxy manually
```

## Development

Run all local checks:

```bash
make check
```

Or run commands directly:

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

## Release

Create and push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will run GoReleaser and upload release artifacts.

## License

MIT
