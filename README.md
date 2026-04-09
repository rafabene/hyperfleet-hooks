# HyperFleet Hooks

Shared validation tools for all HyperFleet projects. This repository is not meant to be cloned or built manually — it is consumed automatically by the [pre-commit](https://pre-commit.com/) framework and by Prow CI via a container image.

## How It Works

### For Developers (pre-commit hook)

Any HyperFleet repository can enforce commit message validation by adding a `.pre-commit-config.yaml` file. The `pre-commit` framework will automatically clone this repository, build the Go binary, cache it, and run it on every commit. No changes to the consuming repository's Makefile or build system are needed.

1. Install the pre-commit framework (one-time):

   ```bash
   pip install pre-commit
   ```

2. Add a `.pre-commit-config.yaml` to your repository:

   ```yaml
   repos:
     - repo: https://github.com/openshift-hyperfleet/hyperfleet-hooks
       rev: v0.1.0  # pin to a specific tag
       hooks:
         - id: hyperfleet-commitlint
   ```

3. Install the hook:

   ```bash
   pre-commit install --hook-type commit-msg
   ```

From this point on, every `git commit` in that repository will validate the commit message automatically. The developer does not need to install Go, clone this repo, or run `make build`.

### For CI (Prow)

In Prow, a pre-built container image (`quay.io/openshift-hyperfleet/hooks`) is used to validate all commits and the PR title. See the [commitlint documentation](docs/commitlint.md) for Prow configuration details.

## Features

### Commitlint

Validates commit messages and PR titles against the [HyperFleet Commit Standard](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md).

**[→ Full Documentation](docs/commitlint.md)**

## Development

These instructions are only needed if you are contributing to this repository.

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linters (requires golangci-lint, managed by bingo)
make image    # Build container image
```

## License

Apache 2.0
