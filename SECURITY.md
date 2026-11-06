# Security Policy

## Supported Versions

NextGate does not currently publish versioned releases. Security fixes are
applied to the latest commit on `main`; older commits and unmaintained forks are
not supported.

| Version | Supported |
| --- | --- |
| Latest `main` | Yes |
| Older commits or forks | No |

## Reporting a Vulnerability

Do not open a public GitHub issue for a suspected vulnerability or include
credentials, tokens, database contents, or other sensitive data in a public
report.

Use GitHub's private vulnerability reporting:

<https://github.com/coderianx/NextGate/security/advisories/new>

If private reporting is unavailable, open a minimal public issue asking for a
private contact channel. Do not include vulnerability details in that issue.

A useful report includes:

- the affected commit or revision,
- a description of the vulnerability and its security impact,
- reproduction steps or a minimal proof of concept,
- the affected endpoint, component, or configuration,
- relevant Go version and operating system, and
- any suggested mitigation or fix.

Reports will be reviewed as soon as possible. Response and remediation time
depends on severity, reproducibility, and maintainer availability. Please allow
time for a fix before publishing technical details.

## Security Model

NextGate currently provides:

- bcrypt hashing for user and administrator passwords;
- HS256 JWTs for access tokens, validated for algorithm, expiry, role, and
  signature;
- random, single-use refresh tokens that are stored only as SHA-256 hashes;
- API keys that are shown once and stored only as SHA-256 hashes;
- AES-256-GCM encryption for provider credentials, with the encryption key held
  in the OS keyring;
- a localhost-only administrator API; and
- SQLite database files created with mode `0600`.

## Deployment Guidance

### Network exposure

The public API binds to `0.0.0.0` by default and serves plain HTTP. Do not expose
it directly to the internet. Use a trusted TLS-terminating reverse proxy,
network firewall rules, request-size limits, and rate limiting in front of the
service. Restrict the admin port to localhost and trusted local processes.

### Secrets

- Use a unique, randomly generated `JWT_SECRET` with at least 32 bytes of
  entropy. Do not reuse a secret from another service.
- Keep `JWT_SECRET`, provider API keys, and issued NextGate API keys out of
  source control, logs, screenshots, and shell history.
- Use a dedicated OS account with an available keyring backend. Protect the
  keyring and the database backup with the same care as the credentials held in
  the database.
- Protect `nextgate.db`, its WAL files, and any backups. The file permissions
  protect the database only while the surrounding directory and host are
  trusted.
- Rotate provider credentials immediately if they may have been exposed. Note
  that changes to `JWT_SECRET` invalidate access tokens, but previously issued
  refresh tokens remain valid until their stored expiry unless the database is
  cleaned separately.

### Administrator access

Run `--admin-password` as the same non-root account that runs NextGate. Do not
use `sudo`: the administrator password hash is stored in the current user's OS
keyring, and a root process cannot access that user's keyring. Do not share the
account context, and do not expose the admin server through a reverse proxy or
public interface. Restrict access to the host and administrator account.

## Known Limitations

- TLS, rate limiting, account lockout, and security-header middleware are not
  built into the servers and must be provided by the deployment environment.
- Expired refresh-token records are not automatically purged, and there is no
  administrative revocation endpoint yet.
- The application has no built-in audit log or alerting for failed logins or
  administrative changes.
- Losing access to the OS keyring makes stored provider credentials
  unrecoverable.

## Scope

Security reports are welcome for issues in this repository, including:

- authentication or authorization bypasses,
- JWT validation or refresh-token flaws,
- cross-user access to API keys or providers,
- SQL injection or unsafe database handling,
- weaknesses in password, API-key, or provider-secret storage,
- secret disclosure through responses or logs, and
- denial-of-service flaws introduced by application code.

Compromise of the host, the OS account, the keyring backend, or a reverse proxy
outside this repository is generally outside the project's control. Reports
should still explain how the application could reduce the resulting impact.
