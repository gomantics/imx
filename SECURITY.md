# Security Policy

## Supported Versions

- `v1.x` (including this branch) — security fixes will be backported where feasible.

## Reporting a Vulnerability

- Email: security@gomantics.com
- Please include a minimal reproducer, affected version/commit, and impact.
- We will acknowledge within 3 business days and provide a timeline for fixes.

## Defensive Defaults

- Default max-bytes limit: 50MB (`WithMaxBytes` to override).
- Streaming buffer capped (`WithBufferSize`).
- Parsers enforce chunk/atom/text limits (PNG, MP4, XMP) to prevent DoS.
- Fuzzing is run regularly across all parsers; add new fuzz cases for new formats.
