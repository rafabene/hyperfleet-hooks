# HyperFleet Hooks

Validation tools for HyperFleet projects.

## Features

### Commitlint

Validates commit messages and PR titles against the [HyperFleet Commit Standard](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md).

**[→ Documentation](docs/commitlint.md)**

## Installation

```bash
# Install pre-commit framework
pip install pre-commit

# Add to your .pre-commit-config.yaml
repos:
  - repo: https://github.com/openshift-hyperfleet/hyperfleet-hooks
    rev: v1.0.0
    hooks:
      - id: hyperfleet-commitlint

# Install hooks (pre-commit will automatically build the binary)
pre-commit install --hook-type commit-msg
```

**Note**: Pre-commit automatically builds the binary. No manual installation needed.

See [commitlint documentation](docs/commitlint.md) for Prow CI setup and detailed usage.

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linters
```

## License

Apache 2.0
