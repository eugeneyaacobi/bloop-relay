# Contributing to bloop-relay

## What maintainers are shipping

This repository ships the server-side relay runtime and its operator-facing release assets.

Changes to top-level docs, examples, and release packaging should keep that positioning clear:
- `README.md` should present this repo as the relay/server repository
- `docs/VPS_DEPLOYMENT.md` should remain the primary operator entry point
- client material in this repo should be described as development or validation support, not the primary public install path
- release-facing copy should avoid transitional wording once behavior is considered the intended release posture

## Branch and release policy

- `main` is the stable release branch.
- All work should land through pull requests.
- Tag stable releases with semver tags in the form `vMAJOR.MINOR.PATCH`.
- Use release notes and top-level docs to describe the current intended operator workflow, not temporary migration language.

## When to use PRs

Use a PR for:
- all feature work
- bug fixes
- docs changes
- workflow/config changes
- dependency changes

Avoid direct pushes to protected branches.

## Docs and release expectations

When changing release-facing docs, verify they stay aligned with the actual workflows in:
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `deploy/docker/relay.Dockerfile`
- `deploy/compose/relay-compose.yml`
- `deploy/examples/relay.example.yaml`

In particular:
- do not describe client installation from this repo as the primary product story
- do not promise artifacts or deployment paths the workflows do not actually publish
- keep operator guidance concrete: server host, proxy, DNS, config, secrets, and verification

## Required protections

Configure GitHub branch protection on `main`:
- require pull requests before merging
- require CI status checks to pass
- require review from code owners
- block force pushes
- block deletions

## Recommended CODEOWNERS scope

Require maintainer review for:
- `.github/workflows/*`
- `deploy/*`
- auth/security-sensitive code
- release automation
- top-level release-facing docs
