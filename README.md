# hyperfleet-hooks

Pre-commit hooks registry for the HyperFleet project ecosystem.

Provides centralized, reusable [pre-commit](https://pre-commit.com/) hooks that enforce
code quality and commit message standards across all HyperFleet repositories.

## Available Hooks

| Hook ID                 | Stage        | Description                                                                                                                                                                |
| ----------------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `hyperfleet-commitlint` | `commit-msg` | Validates commit messages against the [HyperFleet Commit Standard](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md) |

### hyperfleet-commitlint

Enforces the HyperFleet Commit Standard using [commitlint](https://commitlint.js.org/).

#### Accepted formats

```text
HYPERFLEET-123 - feat: add cluster provisioning endpoint

feat: add cluster provisioning endpoint
```

#### Commit types

`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

#### Rules

- Subject line (excluding JIRA prefix) must not exceed 72 characters
- Type must be one of the allowed types listed above
- Subject must not be empty
- Type must be lowercase

## Installation

### Prerequisites

- [pre-commit](https://pre-commit.com/#install) installed
- Node.js 18+ available

### Adding to your repository

Add this to your `.pre-commit-config.yaml`:

```yaml
repos:
  # HyperFleet commit message validation
  - repo: https://github.com/openshift-hyperfleet/hyperfleet-hooks
    rev: v1.0.0 # Use latest release tag
    hooks:
      - id: hyperfleet-commitlint
```

Then install the hooks:

```bash
pre-commit install --hook-type commit-msg
```

## Testing

```bash
make install
make test
```

## Commit Message Format

See the full specification in [commit-standard.md](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md).

### Quick Reference

```text
HYPERFLEET-XXX - <type>: <subject>

[optional body]

[optional footer(s)]
```

#### Breaking Changes

Breaking changes require a `BREAKING CHANGE:` footer:

```text
HYPERFLEET-567 - feat: rename cluster phase to status

BREAKING CHANGE: ClusterStatus.phase field renamed to ClusterStatus.status.
API clients reading cluster status will receive errors on the old field.
Update all references from .phase to .status in your code.
```

## Related

- [HyperFleet Commit Standard](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md)
- [pre-commit](https://pre-commit.com/) — Framework for managing git hooks
