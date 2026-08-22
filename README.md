# p-manager

A terminal-based password manager written in Go. Your vault is encrypted client-side and stored in any S3-compatible cloud storage, so it stays in sync across all of your machines.

## Features

- End-to-end encrypted vault stored in any S3-compatible storage
- Master password protection with key derivation via Argon2id
- Two-factor authentication via 6-digit OTP sent to your email
- Create, view, edit and delete password entries
- Built-in secure password generator
- Clean terminal UI built with Bubble Tea / Bubbles / Lip Gloss
- Secrets (S3 keys, SMTP password) kept in the OS keyring

## Security model

- The vault (`vault.enc`) is encrypted with **AES-256-GCM** before it ever leaves your machine.
- The encryption key is derived from your master password and a random salt using **Argon2id**.
- A verifier, stored in `meta.json`, lets the app confirm you entered the correct master password.
- After the master password is accepted, a one-time code is emailed to you and must be entered to unlock the vault.
- S3 credentials and the SMTP password are stored in the OS keyring (via `go-keyring`), never in the repository.
- The config file (non-secret S3/SMTP settings) is saved locally to your OS config directory.

## Requirements

- Go 1.26 or newer
- Any S3-compatible storage (e.g. Selectel, MinIO, AWS S3) with an access key, secret key, endpoint, region and bucket
- An SMTP account that can send email (used to deliver the one-time codes)

## Installation

```bash
git clone https://github.com/stepan41k/p-manager.git
cd p-manager
go build ./cmd/vault/main.go
```

## Usage

```bash
make run
```

On the first launch the interactive setup wizard asks for:

1. S3 region, endpoint, bucket, access key and secret key
2. SMTP host, port, sender email and sender password
3. Your email (receives the OTP codes)
4. A new master password

After setup, each login requires your master password and the one-time code emailed to you.

## Commands

| Command | Description |
| --- | --- |
| `make run` | Build and run the application |
| `make lint` | Run golangci-lint |

## License

This project is licensed under the [Apache License 2.0](LICENSE).
