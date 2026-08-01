# Releasing

The canonical release process lives in
[RELEASING.md at the repository root](https://github.com/watchthelight/HypergraphGo/blob/main/RELEASING.md).
This page is a summary; when the two disagree, the root document wins.

The short version:

- Releases follow Semantic Versioning.
- Automation is driven by the GitHub `release: published` event. Pushing a
  tag alone does not fire the release workflows; the helper scripts in
  `scripts/` create the tag and the GitHub Release together.
- Preparation happens on `main`: changelog section, release notes file,
  version references, green CI.
- GoReleaser builds both `hg` and `hottgo` for every target, and the
  downstream workflows publish Docker images, Homebrew and Scoop updates,
  Cloudsmith packages, and macOS disk images.

To install a released build, use the
[README installation section](https://github.com/watchthelight/HypergraphGo#installation).
