.PHONY: install test lint

install: ## Install dependencies
	npm install

test: ## Run commitlint validation tests
	@echo "=== Valid commits (should pass) ==="
	@echo "feat: add new feature" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: feat without JIRA prefix" || { echo "FAIL: 'feat: add new feature' should have been accepted"; exit 1; }
	@echo "HYPERFLEET-123 - feat: add new feature" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: feat with JIRA prefix" || { echo "FAIL: 'HYPERFLEET-123 - feat: ...' should have been accepted"; exit 1; }
	@echo "HYPERFLEET-999 - fix: resolve null pointer in cluster handler" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: fix with JIRA prefix" || { echo "FAIL: 'HYPERFLEET-999 - fix: ...' should have been accepted"; exit 1; }
	@echo "docs: update README" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: docs" || { echo "FAIL: 'docs: ...' should have been accepted"; exit 1; }
	@echo "style: run gofmt" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: style" || { echo "FAIL: 'style: ...' should have been accepted"; exit 1; }
	@echo "refactor: extract helper function" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: refactor" || { echo "FAIL: 'refactor: ...' should have been accepted"; exit 1; }
	@echo "perf: optimize query" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: perf" || { echo "FAIL: 'perf: ...' should have been accepted"; exit 1; }
	@echo "test: add unit tests" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: test" || { echo "FAIL: 'test: ...' should have been accepted"; exit 1; }
	@echo "build: update Makefile" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: build" || { echo "FAIL: 'build: ...' should have been accepted"; exit 1; }
	@echo "ci: add GitHub Actions workflow" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: ci" || { echo "FAIL: 'ci: ...' should have been accepted"; exit 1; }
	@echo "chore: update gitignore" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: chore" || { echo "FAIL: 'chore: ...' should have been accepted"; exit 1; }
	@echo "revert: undo last change" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: revert" || { echo "FAIL: 'revert: ...' should have been accepted"; exit 1; }
	@printf "HYPERFLEET-567 - feat: rename cluster phase to status\n\nBREAKING CHANGE: ClusterStatus.phase field renamed to ClusterStatus.status." | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && echo "  ✓ accepted: BREAKING CHANGE footer" || { echo "FAIL: 'BREAKING CHANGE footer' should have been accepted"; exit 1; }
	@echo ""
	@echo "=== Invalid commits (should fail) ==="
	@echo "this is not a valid commit" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && { echo "FAIL: 'no type' should have been rejected"; exit 1; } || echo "  ✓ rejected: no type"
	@echo "invalid: not a valid type" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && { echo "FAIL: 'invalid type' should have been rejected"; exit 1; } || echo "  ✓ rejected: invalid type"
	@echo "feat:" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && { echo "FAIL: 'empty subject' should have been rejected"; exit 1; } || echo "  ✓ rejected: empty subject"
	@echo "feat: this is a very long commit message that exceeds the maximum allowed length of seventy two characters limit" | npx commitlint --config commitlint.config.js > /dev/null 2>&1 && { echo "FAIL: 'header too long' should have been rejected"; exit 1; } || echo "  ✓ rejected: header too long"
	@echo ""
	@echo "=== All tests passed ==="

lint: ## Lint JavaScript files
	npx --yes prettier --check "**/*.js"
