# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-10-05

First release: a CLI and local stdio MCP server covering the standard WHOOP member OAuth API.

### Added

- OAuth 2.0 authorization code flow: `whoopctl auth setup`, `login`, `status`, `refresh`, and `revoke`.
- Automatic access token refresh, with credentials and tokens stored in the OS user config directory.
- Recovery: collection (`whoopctl recovery`) and by cycle (`whoopctl recovery cycle`).
- Cycle: collection (`whoopctl cycle`) and by ID (`whoopctl cycle get`).
- Sleep: collection (`whoopctl sleep`), by ID (`whoopctl sleep get`), and by cycle (`whoopctl sleep cycle`).
- Workout: collection (`whoopctl workout`) and by ID (`whoopctl workout get`).
- Basic user profile (`whoopctl profile`) and body measurement (`whoopctl body`).
- Activity ID mapping from v1 IDs to v2 UUIDs (`whoopctl activity map`).
- Pagination and time filters (`--limit`, `--start`, `--end`, `--next-token`) on collection commands.
- `whoopctl-mcp` local stdio MCP server exposing a tool for every supported endpoint, including `revoke_user_access`.
- `whoopctl version`, reporting the release tag, a Go pseudo-version, or `dev`.

[Unreleased]: https://github.com/Tianbo-Qiu/whoopctl/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Tianbo-Qiu/whoopctl/releases/tag/v0.1.0
