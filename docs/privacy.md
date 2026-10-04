# Privacy Policy

`whoopctl` is a CLI and local MCP server for working with your own WHOOP data.

`whoopctl` does not operate a hosted service. It does not collect, store, sell, or share your personal data.

Your WHOOP client credentials, access tokens, and refresh tokens are stored locally on your machine. API requests are sent directly from your machine to WHOOP.

Do not share your local `whoopctl` config or token files.

## Data Storage

`whoopctl` stores configuration in your operating system's user config directory:

- `whoopctl/config.json` stores your WHOOP app client ID and client secret.
- `whoopctl/token.json` stores your WHOOP OAuth access token, refresh token, token type, scopes, and expiration time.

These files are created with user-only file permissions where supported by the operating system.

## Third Parties

`whoopctl` sends authenticated API requests to WHOOP using the access token created by the OAuth flow. WHOOP's handling of that data is governed by WHOOP's own terms and privacy policies.

## Contact

For questions about this project, open an issue in the `whoopctl` GitHub repository.
