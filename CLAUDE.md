# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

A centralized [pre-commit](https://pre-commit.com/) hooks registry for the HyperFleet project ecosystem. Other HyperFleet repos (api, sentinel, adapter, broker) consume hooks from here via `.pre-commit-config.yaml`.

## Commands

```bash
make install   # Install Node.js dependencies
make test      # Run commitlint validation tests (valid + invalid commit messages)
make lint      # Check JS formatting with prettier
```

To test a single commit message manually:

```bash
echo "feat: my message" | npx commitlint --config commitlint.config.js
```

## Architecture

This repo follows the [pre-commit registry pattern](https://pre-commit.com/#creating-new-hooks). The key flow:

1. `.pre-commit-hooks.yaml` defines hooks that consuming repos reference (entry point for pre-commit framework)
2. `package.json` maps the bin `hyperfleet-commitlint` to `hooks/commitlint-wrapper.js`
3. `hooks/commitlint-wrapper.js` resolves the bundled `commitlint.config.js` and runs commitlint against the commit message file — this indirection is needed because pre-commit runs hooks from the consuming repo's CWD, not from this repo
4. `commitlint.config.js` defines the HyperFleet Commit Standard rules

## Commit Message Standard

Format: `HYPERFLEET-XXX - <type>: <subject>` (JIRA prefix optional).

The custom `headerPattern` regex in `commitlint.config.js` strips the JIRA prefix before parsing. A custom plugin rule (`header-max-length-excluding-jira`) enforces the 72-char limit on the `<type>: <subject>` portion only.

Full spec: [commit-standard.md](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md)

## Related Repos

- [architecture](https://github.com/openshift-hyperfleet/architecture) — Standards and architecture docs
