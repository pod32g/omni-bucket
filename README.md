# bb — Bitbucket Cloud CLI (omni-bucket)

A command-line client for Bitbucket Cloud (REST API 2.0).

## Install

```bash
go build -o bb ./cmd/omni-bucket
```

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

## Usage

```bash
bb pr list --repo workspace/repo [--state open|merged|declined] [--limit N] [--json]
bb pr view <id> --repo workspace/repo
bb pr approve <id> --repo workspace/repo
bb pr merge <id> --repo workspace/repo
bb pr create --repo workspace/repo --title "..." --source feature/x [--destination main]

bb repo list --workspace my-workspace [--json]
bb pipeline list --repo workspace/repo [--json]
bb issue list --repo workspace/repo [--json]
```

Every command supports `--json` for scripting.
