# NextGate

NextGate is a self-hosted, local-first API gateway for managing AI provider
credentials and issuing API keys to authenticated users. It exposes a public
HTTP API for user authentication and API-key management, plus a separate
administrator API for configuring providers.

> **Project status:** NextGate is under active development. The current code
> implements authentication, API-key management, and provider configuration.
> Model-proxy or completion routes are not implemented yet.

## Features

- Separate public and administrator HTTP servers.
- User registration and login with bcrypt password hashing.
- Short-lived HS256 JWT access tokens and rotating refresh tokens.
- API keys issued to users and stored only as SHA-256 hashes.
- Administrator-managed AI providers with built-in defaults for Groq.
- Provider API keys encrypted at rest with AES-256-GCM.
- OS-keyring-backed encryption and administrator password storage.
- SQLite storage with WAL mode, foreign keys, a busy timeout, secure deletion,
  and restrictive file permissions.

## Architecture

| Component | Default address | Purpose |
| --- | --- | --- |
| Public API | `0.0.0.0:8080` | User authentication and API-key management |
| Admin API | `127.0.0.1:9090` | Administrator login and provider management |
| SQLite | `./nextgate.db` | Users, refresh tokens, providers, and API keys |

The public API binds to all interfaces by default and serves plain HTTP. Put it
behind a TLS-terminating reverse proxy and a firewall before exposing it outside
the local machine. The admin API binds to loopback only.

## Requirements

- Go 1.27.1 or newer.
- A working OS keyring backend:
  - Secret Service or KWallet on Linux,
  - Keychain on macOS,
  - Windows Credential Manager on Windows.
- A random `JWT_SECRET` value of at least 32 bytes.

SQLite is provided by the pure-Go `modernc.org/sqlite` driver; no separate
SQLite installation is required.

## Quick Start

```sh
git clone https://github.com/coderianx/NextGate.git
cd NextGate

go build -o nextgate .

export JWT_SECRET="$(openssl rand -base64 48)"
sudo -E ./nextgate --admin-password
```

`--admin-password` prompts for a new administrator password and stores a bcrypt
hash in the OS keyring. With `JWT_SECRET` available, it then starts both servers.
The flag currently requires root. The keyring entry belongs to the account
context in which the process runs, so later runs must use the same
account/keyring context.

Subsequent runs use the same environment and keyring context:

```sh
export JWT_SECRET="$(cat /path/to/jwt-secret)"
sudo -E ./nextgate
```

The current startup path reads `JWT_SECRET` from the process environment. It
does not load an environment file automatically. Keep the secret out of shell
history and source control, for example by loading it from a protected secret
manager or environment file.

Changing ports:

```sh
./nextgate --api-port 8080 --admin-port 9090
```

## Configuration

| Variable or flag | Required | Default | Description |
| --- | --- | --- | --- |
| `JWT_SECRET` | Yes | — | Secret used to sign and verify HS256 access tokens |
| `--api-port` | No | `8080` | Public API port; the server binds to `0.0.0.0` |
| `--admin-port` | No | `9090` | Admin API port; the server binds to `127.0.0.1` |
| `--admin-password` | No | `false` | Set the administrator password; requires root |

The SQLite database is currently created as `nextgate.db` in the process
working directory.

Generate a suitable JWT secret with:

```sh
openssl rand -base64 48
```

## API Overview

All request and response bodies use JSON. Authenticated endpoints expect an
`Authorization: Bearer <token>` header.

### Public API

| Method | Path | Authentication | Description |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/register` | None | Register a user |
| `POST` | `/api/v1/auth/login` | None | Log in and receive an access/refresh token pair |
| `POST` | `/api/v1/auth/refresh` | Refresh token in body | Rotate the refresh token and issue a new access token |
| `POST` | `/api/v1/apikeys` | User JWT | Create an API key |
| `GET` | `/api/v1/apikeys` | User JWT | List the user's API keys |
| `DELETE` | `/api/v1/apikeys/{id}` | User JWT | Delete one of the user's API keys |

Passwords must not exceed 72 bytes. Access tokens expire after 60 minutes.
Refresh tokens expire after 30 days and are rotated on every successful refresh.

### Admin API

| Method | Path | Authentication | Description |
| --- | --- | --- | --- |
| `POST` | `/api/v1/login` | None | Log in as administrator |
| `POST` | `/api/v1/providers` | Admin JWT | Add a provider |
| `GET` | `/api/v1/providers` | Admin JWT | List configured providers |
| `DELETE` | `/api/v1/providers/{id}` | Admin JWT | Delete a provider |

Provider names are passed through as supplied. The name `groq` always uses
`https://api.groq.com/openai/v1`; any supplied `base_url` is replaced. Other
providers must include a `base_url`.

### Example: User Login and API-Key Creation

Register a user:

```sh
curl -sS -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"replace-with-a-strong-password"}'
```

Log in:

```sh
curl -sS -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"replace-with-a-strong-password"}'
```

Create an API key using the returned `accessToken`:

```sh
curl -sS -X POST http://127.0.0.1:8080/api/v1/apikeys \
  -H 'Authorization: Bearer <accessToken>' \
  -H 'Content-Type: application/json' \
  -d '{"title":"development"}'
```

The generated API key is returned only once. NextGate stores its SHA-256 hash,
not the key itself.

### Example: Add an Admin Provider

```sh
curl -sS -X POST http://127.0.0.1:9090/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"password":"<admin-password>"}'

curl -sS -X POST http://127.0.0.1:9090/api/v1/providers \
  -H 'Authorization: Bearer <adminAccessToken>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"groq","api_key":"<groq-api-key>"}'
```

## Data Protection

- User and administrator passwords are hashed with bcrypt (cost 12).
- Refresh tokens are random 256-bit values stored as SHA-256 hashes.
- Issued API keys use the `sk-ng-v1-` prefix and are stored as SHA-256 hashes.
- Provider credentials are encrypted with AES-256-GCM before being written to
  SQLite. The encryption key is stored in the OS keyring.
- The database file is created with mode `0600` and configured with WAL,
  foreign-key enforcement, and secure deletion.

See [SECURITY.md](SECURITY.md) for vulnerability reporting and deployment
guidance.

## Development

```sh
go build ./...
go test ./...
go vet ./...
go fmt ./...
```

There are currently no test files in the repository. The commands above are the
expected baseline checks for new contributions.

## Project Layout

```text
.
├── main.go                                  # Process entry point and HTTP servers
├── internal/
│   ├── admin/                              # Admin API, login, and providers
│   ├── api/                                # Public API, auth, and API keys
│   ├── config/                             # Environment and keyring configuration
│   ├── middlewares/                        # JWT authentication middleware
│   ├── models/                             # Request and response models
│   ├── secret/                             # AES-256-GCM encryption
│   └── store/                              # SQLite connection and schema
├── go.mod
└── LICENSE
```

## License

NextGate is released under the [MIT License](LICENSE).
