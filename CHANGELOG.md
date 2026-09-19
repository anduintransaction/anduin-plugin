# Changelog

Versions follow [semantic versioning](https://semver.org). Every release `X.Y.Z` is published as the tag `vX.Y.Z` (Production US) and `vX.Y.Z-eu` (Production EU); the two builds differ only in the server they connect to.

## 0.10.0

- Production EU support: each release is also published as a `vX.Y.Z-eu` tag that connects to `https://mcp.eu.anduin.app/mcp`, and the marketplace gains an `anduin-eu` entry. Install one region only.
- Organization admins choose the region with the `ref` of their marketplace entry; see the admin guide in `docs/`.
- GitHub Releases list the paste-ready `ref` lines, the matching commits, and one ZIP per region.
- Local testing uses `claude --plugin-dir`.

## Earlier releases

See the git history and tags.
