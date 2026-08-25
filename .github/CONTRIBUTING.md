# Contributing to p-manager

Thanks for your interest in contributing! This document describes how to set up the project, submit changes, and the rules that keep p-manager secure and maintainable.

## Reporting security vulnerabilities

**Do not open public issues for security vulnerabilities.**

p-manager is a password manager, so responsible disclosure matters. If you find a vulnerability:

1. Use [GitHub's private vulnerability reporting](https://github.com/stepan41k/p-manager/security/advisories/new), or
2. Contact the maintainer directly.

Include steps to reproduce, affected versions, and your assessment of the impact. You will be credited unless you prefer to remain anonymous.

For non-security bugs and feature requests, please open a regular GitHub issue.

## Development setup

### Prerequisites

- Go **1.26 or newer**
- [golangci-lint](https://golangci-lint.run/) (for `make lint`)
- An S3-compatible storage bucket and an SMTP account (needed to actually run the app end-to-end)

### Build and run

```bash
git clone https://github.com/stepan41k/p-manager.git
cd p-manager

make run      # run the TUI locally
make version  # verify the binary builds and reports its version
```

On first launch the app walks you through an interactive setup (S3 credentials, SMTP settings, master password).

### Tests and linting

```bash
go test ./...     # run all tests
make lint         # run golangci-lint
go build ./...    # make sure everything compiles
```

CI runs on every release tag; run these commands locally before pushing.

## Project structure

```
cmd/vault/            CLI entry point
internal/app/         Bubble Tea TUI (model/update/view, forms, devices, email OTP)
internal/config/      Configuration loading & saving (config.json)
internal/crypto/      AES-256-GCM, Argon2id/HKDF, OTP, password generation
internal/lib/logger/  Logging helpers
internal/storage/s3/  S3-compatible storage client & keyring credentials
internal/sys/         Platform-specific hardening (memory dump protection)
```

Keep new code inside `internal/` unless it genuinely needs to be importable from outside.

## Submitting changes

1. Fork the repository and create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes.
3. Add or update tests for any changed behavior.
4. Run `go test ./...`, `go build ./...`, and `make lint` — all must pass.
5. Open a pull request with a clear description of *what* and *why*.

Small, focused PRs are merged faster than large ones.

### Commit messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/) with package scopes, for example:

```
feat(app/update): added update for settings state
fix(crypto): use rejection sampling in GeneratePassword
refactor(app/devices): changed revokeAllDevices function
docs(readme): updated key bindings table
```

Common scopes: `app/<file>`, `crypto`, `config`, `storage/s3`.

## Code guidelines

- Follow the style of the surrounding code; `make lint` is the source of truth.
- Return errors explicitly instead of ignoring them (`_ =` requires justification).
- Keep secrets out of logs — never log passwords, keys, tokens, or vault contents.
- User-facing strings are written in English.
- Prefer small helper functions over duplication across `update.go` / `view.go`.

## Security-sensitive contributions

Extra rules apply because this project handles secrets:

- **Never commit real credentials**: S3 keys, SMTP passwords, master passwords, device tokens. The app stores these in the OS keyring; there is nothing to put into files.
- Changes to `internal/crypto/` require:
  - A clear explanation of the cryptographic rationale.
  - Tests covering edge cases (empty input, short data, wrong length).
  - Extra reviewer scrutiny — expect questions.
- Any change affecting how the vault, metadata, or trusted devices are stored must remain backward-compatible with existing encrypted `vault.enc` / `meta.json` files, or include a migration path.
- Do not weaken defaults (KDF parameters, OTP limits, clipboard/inactivity timeouts) without discussing it in an issue first.

## Release process (maintainers)

Releases are automated via GitHub Actions:

1. Push a tag `vX.Y.Z`.
2. GoReleaser builds binaries for Linux/macOS/Windows (amd64/arm64), publishes the release with checksums, updates the Homebrew tap.
3. A separate job builds the Windows installer (`installer.iss`) and attaches it to the release.

Release notes are generated from conventional commits into `RELEASE_NOTES.md`.

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE) that covers this project.
