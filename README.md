# bb — Bitbucket Cloud CLI (omni-bucket)

A command-line client for Bitbucket Cloud (REST API 2.0).

## Install

```bash
make build                       # builds ./bb stamped with the git version
# or a plain build (version comes from embedded VCS info):
go build -o bb ./cmd/omni-bucket
```

Check the version:

```bash
bb --version                     # e.g. "bb v0.1.0 (a1b2c3d4e5f6, 2026-06-10T12:00:00Z)"
```

`make build` injects the version from `git describe --tags --always --dirty`. A
plain `go build` still reports the commit and build date from the Go toolchain's
embedded VCS metadata.

## Authenticate

`bb` uses a scoped Bitbucket Cloud **API token** (Basic auth with your email).
Create one at: Bitbucket → Personal settings → API tokens → *Create API token with scopes*.

```bash
bb auth login            # prompts for email + token (stored in ~/.config/omni-bucket/config.yml, 0600)
bb auth status           # show the authenticated account
```

Or use environment variables (these override the config file):

```bash
export BITBUCKET_EMAIL="you@example.com"
export BITBUCKET_API_TOKEN="<scoped token>"
export BITBUCKET_WORKSPACE="my-workspace"   # optional default workspace
```

### Browser login (OAuth 2.0)

`bb` can also log in through your browser using a Bitbucket OAuth consumer you
create once:

1. In Bitbucket: **Workspace settings → OAuth consumers → Add consumer**.
   - **Callback URL:** `http://127.0.0.1:8765/callback`
   - **Permissions (scopes):** grant what you need — Account, Repositories,
     Pull requests, Pipelines, Issues (read/write as appropriate).
2. Run:
   ```bash
   bb auth login --browser
   ```
   You'll be prompted for the consumer **Key** (`client_id`) and **Secret** the
   first time (or pass `--client-id`/`--client-secret`, or set
   `BITBUCKET_OAUTH_CLIENT_ID` / `BITBUCKET_OAUTH_CLIENT_SECRET`). `bb` opens the
   browser, you approve, and the tokens are stored in
   `~/.config/omni-bucket/config.yml` (0600) and refreshed automatically.

Remove stored credentials at any time:

```bash
bb auth logout
```

## Usage

```bash
bb pr list --repo workspace/repo [--state open|merged|declined|superseded] [--limit N] [--json]
bb pr view <id> --repo workspace/repo
bb pr approve <id> --repo workspace/repo
bb pr merge <id> --repo workspace/repo
bb pr create --repo workspace/repo --title "..." --source feature/x [--destination main]

bb repo list --workspace my-workspace [--json]
bb pipeline list --repo workspace/repo [--json]
bb issue list --repo workspace/repo [--json]
```

All `list` and `view` commands, plus `pr approve`/`merge`/`create`, support `--json` for scripting.
