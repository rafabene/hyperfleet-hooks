package commitlint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/commitlint"
	ghclient "github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/github"
	"github.com/spf13/cobra"
)

var (
	// ErrStopIteration is used to stop commit iteration when base is reached
	ErrStopIteration = errors.New("stop iteration")

	// shaPattern matches valid git SHA (7-40 hex characters)
	shaPattern = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
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
	// Open git repository once and reuse
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get commit range from environment
	baseSHA, headSHA, err := getCommitRange(repo)
	if err != nil {
		return fmt.Errorf("failed to get commit range: %w", err)
	}

	// Validate SHA format
	if err := validateSHA(baseSHA); err != nil {
		return fmt.Errorf("invalid base SHA: %w", err)
	}
	if err := validateSHA(headSHA); err != nil {
		return fmt.Errorf("invalid head SHA: %w", err)
	}

	fmt.Fprintf(os.Stderr, "🔍 Validating commits in range: %s..%s\n", baseSHA[:8], headSHA[:8])

	// Get all commits in range using git log
	commits, err := getCommitsInRange(repo, baseSHA, headSHA)
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
		msg, subject, err := getCommitMessage(repo, sha)
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	title, err := client.GetPRTitleFromEnv(ctx)
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
func getCommitRange(repo *git.Repository) (baseSHA, headSHA string, err error) {
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
	baseRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", baseBranch), true)
	if err != nil {
		return "", "", fmt.Errorf("failed to get base SHA: %w", err)
	}
	baseSHA = baseRef.Hash().String()

	// Get SHA of HEAD
	head, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("failed to get HEAD SHA: %w", err)
	}
	headSHA = head.Hash().String()

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

// validateSHA checks if a string is a valid git SHA
func validateSHA(sha string) error {
	if sha == "" {
		return fmt.Errorf("SHA cannot be empty")
	}
	if !shaPattern.MatchString(sha) {
		return fmt.Errorf("invalid SHA format: %s (expected 7-40 hex characters)", sha)
	}
	return nil
}

// getCommitsInRange returns all commit SHAs in the given range
func getCommitsInRange(repo *git.Repository, baseSHA, headSHA string) ([]string, error) {
	// Get commit objects
	headHash := plumbing.NewHash(headSHA)
	baseHash := plumbing.NewHash(baseSHA)

	// Create log iterator from head
	commitIter, err := repo.Log(&git.LogOptions{
		From: headHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	var commits []string
	err = commitIter.ForEach(func(c *object.Commit) error {
		// Stop when we reach the base commit
		if c.Hash == baseHash {
			return ErrStopIteration
		}
		commits = append(commits, c.Hash.String())
		return nil
	})

	// ErrStopIteration is expected when we reach the base commit
	if err != nil && !errors.Is(err, ErrStopIteration) {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// getCommitMessage returns the full message and subject of a commit
func getCommitMessage(repo *git.Repository, sha string) (message, subject string, err error) {
	// Get commit object
	hash := plumbing.NewHash(sha)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return "", "", fmt.Errorf("failed to get commit: %w", err)
	}

	// Full message
	message = strings.TrimSpace(commit.Message)

	// Subject is the first line of the message
	lines := strings.SplitN(message, "\n", 2)
	subject = strings.TrimSpace(lines[0])

	return message, subject, nil
}
