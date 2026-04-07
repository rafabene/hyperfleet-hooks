package commitlint

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/commitlint"
	ghclient "github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/github"
	"github.com/spf13/cobra"
)

var (
	pr bool
)

// NewCommand creates the commitlint command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitlint [commit-msg-file]",
		Short: "Validate commit message",
		Long: `Validate commit messages against HyperFleet Commit Standard.

Usage:
  # Local: Validate single commit message
  hyperfleet-hooks commitlint <file>           # from file (e.g., .git/COMMIT_EDITMSG)
  echo "feat: ..." | hyperfleet-hooks commitlint  # from stdin

  # CI: Validate entire PR (all commits + PR title)
  hyperfleet-hooks commitlint --pr

Examples:
  # Pre-commit hook (local)
  hyperfleet-hooks commitlint .git/COMMIT_EDITMSG

  # Manual validation (local)
  echo "feat: add new feature" | hyperfleet-hooks commitlint

  # Prow CI (auto-detects PR from environment)
  hyperfleet-hooks commitlint --pr`,
		Args: cobra.MaximumNArgs(1),
		RunE: run,
	}

	cmd.Flags().BoolVar(&pr, "pr", false, "Validate all commits and PR title (CI mode, auto-detect from environment)")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	validator := commitlint.NewValidator()

	// If --pr is provided, validate all commits and PR title
	if pr {
		return validatePR(validator)
	}

	// Single commit validation (from file or stdin)
	var message string
	var err error

	if len(args) > 0 {
		// Read from file
		filePath := args[0]
		result, err := validator.ValidateFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read commit message: %w", err)
		}
		return handleValidationResult(result, filePath)
	}

	// Read from stdin
	content, err := readStdin()
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}
	message = content

	result := validator.Validate(message)
	return handleValidationResult(result, "stdin")
}

func handleValidationResult(result *commitlint.ValidationResult, source string) error {
	if result.Valid {
		return nil
	}

	// Print errors in commitlint format
	fmt.Fprintf(os.Stderr, "⧗   input: %s\n", source)
	for _, err := range result.Errors {
		fmt.Fprintf(os.Stderr, "✖   %s\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "\n✖   found %d problems, 0 warnings\n", len(result.Errors))
	fmt.Fprintln(os.Stderr, "ⓘ   Get help: https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md")

	return fmt.Errorf("validation failed")
}

func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("no input provided (stdin is empty)")
	}

	content, err := os.ReadFile("/dev/stdin")
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func validatePR(validator *commitlint.Validator) error {
	// Get commit range from environment
	baseSHA, headSHA, err := getCommitRange()
	if err != nil {
		return fmt.Errorf("failed to get commit range: %w", err)
	}

	fmt.Fprintf(os.Stderr, "🔍 Validating commits in range: %s..%s\n", baseSHA[:8], headSHA[:8])

	// Get all commits in range using git log
	commits, err := getCommitsInRange(baseSHA, headSHA)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️  No commits found in range")
		return nil
	}

	fmt.Fprintf(os.Stderr, "📝 Found %d commit(s) to validate\n\n", len(commits))

	// Validate each commit
	var failedCommits []string
	passedCount := 0

	for _, sha := range commits {
		// Get commit message using git show
		msg, subject, err := getCommitMessage(sha)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to get message for %s: %v\n", sha[:8], err)
			failedCommits = append(failedCommits, sha[:8])
			continue
		}

		fmt.Fprintf(os.Stderr, "Checking: %s - %s\n", sha[:8], subject)

		result := validator.Validate(msg)
		if result.Valid {
			fmt.Fprintln(os.Stderr, "  ✅ PASS")
			passedCount++
		} else {
			fmt.Fprintln(os.Stderr, "  ❌ FAIL")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  Error details:")
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "    ✖ %s\n", e.Error())
			}
			fmt.Fprintln(os.Stderr, "")
			failedCommits = append(failedCommits, sha[:8])
		}
	}

	// Validate PR title
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintln(os.Stderr, "📋 Validating PR title...")

	prTitleFailed := false
	client := ghclient.NewClient()
	title, err := client.GetPRTitleFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not fetch PR title: %v\n", err)
		fmt.Fprintln(os.Stderr, "   Skipping PR title validation")
	} else {
		prNumberStr := os.Getenv("PULL_NUMBER")
		fmt.Fprintf(os.Stderr, "PR #%s: %s\n", prNumberStr, title)

		result := validator.ValidatePRTitle(title)
		if result.Valid {
			fmt.Fprintln(os.Stderr, "  ✅ PASS")
		} else {
			fmt.Fprintln(os.Stderr, "  ❌ FAIL")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  Error details:")
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "    ✖ %s\n", e.Error())
			}
			fmt.Fprintln(os.Stderr, "")
			prTitleFailed = true
		}
	}

	// Summary
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(failedCommits) == 0 && !prTitleFailed {
		fmt.Fprintf(os.Stderr, "✅ All %d commit(s) passed validation!\n", passedCount)
		if title != "" {
			fmt.Fprintln(os.Stderr, "✅ PR title passed validation!")
		}
		fmt.Fprintln(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return nil
	}

	if len(failedCommits) > 0 {
		fmt.Fprintf(os.Stderr, "❌ %d of %d commit(s) failed validation\n", len(failedCommits), len(commits))
	}
	if prTitleFailed {
		fmt.Fprintln(os.Stderr, "❌ PR title failed validation")
	}
	fmt.Fprintln(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "📖 HyperFleet Commit Standard:")
	fmt.Fprintln(os.Stderr, "   Format: [HYPERFLEET-XXX - ]<type>: <subject>")
	fmt.Fprintln(os.Stderr, "   Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
	fmt.Fprintln(os.Stderr, "   Max length: 72 chars (excluding JIRA prefix)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "   Full spec: https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/standards/commit-standard.md")

	return fmt.Errorf("validation failed")
}

// getCommitRange returns the base and head SHA for the PR commits
// Priority: PULL_REFS > PULL_BASE_SHA+PULL_PULL_SHA > PULL_BASE_REF+HEAD
func getCommitRange() (baseSHA, headSHA string, err error) {
	// Priority 1: PULL_REFS (most accurate, Prow standard)
	if pullRefs := os.Getenv("PULL_REFS"); pullRefs != "" {
		baseSHA, headSHA, err = parsePullRefs(pullRefs)
		if err == nil {
			return baseSHA, headSHA, nil
		}
		fmt.Fprintf(os.Stderr, "⚠️  Failed to parse PULL_REFS: %v\n", err)
	}

	// Priority 2: PULL_BASE_SHA + PULL_PULL_SHA
	baseSHA = os.Getenv("PULL_BASE_SHA")
	headSHA = os.Getenv("PULL_PULL_SHA")
	if baseSHA != "" && headSHA != "" {
		return baseSHA, headSHA, nil
	}

	// Priority 3: PULL_BASE_REF + HEAD (fallback for local testing)
	baseBranch := os.Getenv("PULL_BASE_REF")
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Get SHA of origin/baseBranch
	cmd := exec.Command("git", "rev-parse", "origin/"+baseBranch)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get base SHA: %w", err)
	}
	baseSHA = strings.TrimSpace(string(output))

	// Get SHA of HEAD
	cmd = exec.Command("git", "rev-parse", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get HEAD SHA: %w", err)
	}
	headSHA = strings.TrimSpace(string(output))

	return baseSHA, headSHA, nil
}

// parsePullRefs parses PULL_REFS environment variable
// Format: "base:base_sha,pr_number:pr_sha[,pr_number:pr_sha...]"
// Example: "main:abc123,456:def789"
func parsePullRefs(pullRefs string) (baseSHA, prSHA string, err error) {
	if pullRefs == "" {
		return "", "", fmt.Errorf("PULL_REFS is empty")
	}

	// Split by comma: ["base:base_sha", "pr_number:pr_sha", ...]
	refs := strings.Split(pullRefs, ",")
	if len(refs) < 2 {
		return "", "", fmt.Errorf("invalid PULL_REFS format: %s", pullRefs)
	}

	// Parse base ref (first element)
	baseParts := strings.Split(refs[0], ":")
	if len(baseParts) != 2 {
		return "", "", fmt.Errorf("invalid base ref: %s", refs[0])
	}
	baseSHA = baseParts[1]

	// Parse PR ref (second element, first PR)
	prParts := strings.Split(refs[1], ":")
	if len(prParts) != 2 {
		return "", "", fmt.Errorf("invalid PR ref: %s", refs[1])
	}
	prSHA = prParts[1]

	return baseSHA, prSHA, nil
}

// getCommitsInRange returns all commit SHAs in the given range
func getCommitsInRange(baseSHA, headSHA string) ([]string, error) {
	commitRange := fmt.Sprintf("%s..%s", baseSHA, headSHA)
	cmd := exec.Command("git", "log", "--format=%H", commitRange)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commits) == 1 && commits[0] == "" {
		return []string{}, nil
	}

	return commits, nil
}

// getCommitMessage returns the full message and subject of a commit
func getCommitMessage(sha string) (message, subject string, err error) {
	// Get full message
	cmd := exec.Command("git", "log", "-1", "--format=%B", sha)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit message: %w", err)
	}
	message = strings.TrimSpace(string(output))

	// Get subject
	cmd = exec.Command("git", "log", "-1", "--format=%s", sha)
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit subject: %w", err)
	}
	subject = strings.TrimSpace(string(output))

	return message, subject, nil
}
