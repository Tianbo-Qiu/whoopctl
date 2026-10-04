# whoopctl

<p>
    <img src="docs/assets/logo.webp" alt="whoopctl logo" width="160">
</p>

![CLI](https://img.shields.io/badge/CLI-whoopctl-111827)
![MCP](https://img.shields.io/badge/MCP-local%20stdio-10b981)
![WHOOP](https://img.shields.io/badge/WHOOP-OAuth%202.0-6366f1)

`whoopctl` is a CLI and local MCP server for working with your own WHOOP data. It stores credentials locally and calls the WHOOP API directly from your machine.

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
| Recovery by cycle |  |  |
| Cycle collection |  |  |
| Cycle by ID |  |  |
| Sleep collection |  |  |
| Sleep by ID |  |  |
| Sleep by cycle |  |  |
| Workout collection |  |  |
| Workout by ID |  |  |
| Basic user profile |  |  |
| Body measurement |  |  |
| Revoke user access |  |  |
| Activity ID mapping |  |  |
