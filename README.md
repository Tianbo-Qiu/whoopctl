# whoopctl

<p>
    <img src="docs/assets/logo.webp" alt="whoopctl logo" width="160">
</p>

![CLI](https://img.shields.io/badge/CLI-whoopctl-111827)
![MCP](https://img.shields.io/badge/MCP-local%20stdio-10b981)
![WHOOP](https://img.shields.io/badge/WHOOP-OAuth%202.0-6366f1)

`whoopctl` is a CLI and local MCP server for working with your own WHOOP data. It stores credentials locally and calls the WHOOP API directly from your machine.

## Install

Build the CLI and MCP server binaries from the repo:

```sh
go build -o ./bin/whoopctl ./cmd/whoopctl
go build -o ./bin/whoopctl-mcp ./cmd/whoopctl-mcp
```

You can also install them into your Go binary directory:

```sh
go install ./cmd/whoopctl
go install ./cmd/whoopctl-mcp
```

Make sure your Go binary directory is on your `PATH` if you use `go install`.

## Setup

`whoopctl` uses WHOOP's OAuth 2.0 authorization code flow. You bring your own WHOOP Developer app credentials, log in with your WHOOP account, approve scopes, and `whoopctl` stores user access tokens locally.

### 1. Create a WHOOP Developer app

Create an app in the [WHOOP Developer Dashboard](https://developer-dashboard.whoop.com/).

Fill in the app details WHOOP asks for. For a personal/sandbox app, the privacy policy can be a simple project page:

- Privacy Policy URL: `https://tianbo-qiu.github.io/whoopctl/privacy`

This is app metadata for WHOOP registration. The OAuth redirect URL is separate and must match `whoopctl` exactly.

Set the redirect URL to:

```text
http://127.0.0.1:1061/callback
```

WHOOP may also call this a callback URL. It must match exactly because `whoopctl auth login` starts a local callback server on that address.

Add the scopes you want `whoopctl` to request:

```text
offline
read:recovery
read:cycles
read:sleep
read:workout
read:profile
read:body_measurement
```

`offline` is required for refresh tokens, so `whoopctl` can refresh your access token after it expires.

### 2. Save your app credentials

Copy the app's client ID and client secret from the WHOOP Developer Dashboard, then run:

```sh
whoopctl auth setup \
  --client-id YOUR_CLIENT_ID \
  --client-secret YOUR_CLIENT_SECRET
```

Credentials are stored locally in your OS user config directory under `whoopctl/config.json`.

### 3. Authorize your WHOOP account

Run:

```sh
whoopctl auth login
```

The command prints an authorization URL. Open it in your browser, approve access, and WHOOP will redirect back to the local callback URL.

After login, tokens are stored locally in your OS user config directory under `whoopctl/token.json`.

You can check the current auth state with:

```sh
whoopctl auth status
```

And manually refresh the access token with:

```sh
whoopctl auth refresh
```

To revoke `whoopctl`'s access to your WHOOP data and remove the local token, run:

```sh
whoopctl auth revoke
```

Run `whoopctl auth login` again to re-authorize afterwards.

## CLI examples

| Roadmap item | Command |
| --- | --- |
| Recovery collection | `whoopctl recovery --limit 1` |
| Recovery by cycle | `whoopctl recovery cycle CYCLE_ID` |
| Cycle collection | `whoopctl cycle --limit 1` |
| Cycle by ID | `whoopctl cycle get CYCLE_ID` |
| Sleep collection | `whoopctl sleep --limit 1` |
| Sleep by ID | `whoopctl sleep get SLEEP_ID` |
| Sleep by cycle | `whoopctl sleep cycle CYCLE_ID` |
| Workout collection | `whoopctl workout --limit 1` |
| Workout by ID | `whoopctl workout get WORKOUT_ID` |
| Basic user profile | `whoopctl profile` |
| Body measurement | `whoopctl body` |
| Revoke user access | `whoopctl auth revoke` |
| Activity ID mapping | `whoopctl activity map V1_ACTIVITY_ID` |

Paginated collection commands accept the same filtering flags:

```sh
whoopctl cycle \
  --limit 25 \
  --start START_RFC3339 \
  --end END_RFC3339
```

If a response includes `next_token`, pass it to fetch the next page:

```sh
whoopctl recovery --next-token NEXT_TOKEN
```

## MCP server

`whoopctl-mcp` runs a local stdio MCP server. Configure your MCP client to launch the binary:

```json
{
  "mcpServers": {
    "whoopctl": {
      "command": "/path/to/whoopctl-mcp"
    }
  }
}
```

After connecting, the server exposes tools for supported WHOOP endpoints, including `get_recovery`, `get_recovery_for_cycle`, `get_cycle`, `get_cycle_by_id`, `get_sleep`, `get_sleep_by_id`, `get_sleep_for_cycle`, `get_workout`, `get_workout_by_id`, `get_profile_basic`, `get_body_measurement`, `get_activity_mapping`, and `revoke_user_access`.

## Roadmap

This roadmap tracks the standard WHOOP OAuth API surface for a personal `whoopctl` app.

Auth is handled by the CLI and shared by the local MCP server through the same local config and token files.

| Auth command | CLI |
| --- | --- |
| Setup app credentials | ✅ |
| Login with WHOOP | ✅ |
| Refresh token | ✅ |

> [!NOTE]
> `whoopctl` supports the standard WHOOP member OAuth API. Trusted Partner endpoints are not supported because WHOOP documents them as a separate healthcare partner API for approved partners only. Those endpoints use the OAuth 2.0 client credentials grant with partner credentials, without a WHOOP member login.

| Name | CLI | MCP |
| --- | --- | --- |
| Recovery collection | ✅ | ✅ |
| Recovery by cycle | ✅ | ✅ |
| Cycle collection | ✅ | ✅ |
| Cycle by ID | ✅ | ✅ |
| Sleep collection | ✅ | ✅ |
| Sleep by ID | ✅ | ✅ |
| Sleep by cycle | ✅ | ✅ |
| Workout collection | ✅ | ✅ |
| Workout by ID | ✅ | ✅ |
| Basic user profile | ✅ | ✅ |
| Body measurement | ✅ | ✅ |
| Revoke user access | ✅ | ✅ |
| Activity ID mapping | ✅ | ✅ |
