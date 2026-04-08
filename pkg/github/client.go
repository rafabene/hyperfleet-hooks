package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
)

// Client wraps GitHub API client
type Client struct {
	client *github.Client
}

// NewClient creates a new GitHub client
// It will use GITHUB_TOKEN environment variable if available
func NewClient() *Client {
	var client *github.Client
	token := os.Getenv("GITHUB_TOKEN")

	if token != "" {
		client = github.NewClient(nil).WithAuthToken(token)
	} else {
		client = github.NewClient(nil)
	}

	return &Client{
		client: client,
	}
}

// GetPRTitle fetches the title of a pull request
// repo format: "owner/name" or "org/repo"
func (c *Client) GetPRTitle(ctx context.Context, repo string, prNumber int) (string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo format: %s (expected owner/repo)", repo)
	}

	owner := parts[0]
	repoName := parts[1]

	pr, _, err := c.client.PullRequests.Get(ctx, owner, repoName, prNumber)
	if err != nil {
		return "", fmt.Errorf("failed to fetch PR #%d: %w", prNumber, err)
	}

	if pr.Title == nil {
		return "", fmt.Errorf("PR #%d has no title", prNumber)
	}

	return *pr.Title, nil
}

// GetPRTitleFromEnv fetches PR title using environment variables
// Uses REPO_OWNER, REPO_NAME, PULL_NUMBER from Prow
func (c *Client) GetPRTitleFromEnv(ctx context.Context) (string, error) {
	owner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	prNumberStr := os.Getenv("PULL_NUMBER")

	if owner == "" || repoName == "" || prNumberStr == "" {
		return "", fmt.Errorf("missing required environment variables (REPO_OWNER, REPO_NAME, PULL_NUMBER)")
	}

	var prNumber int
	_, err := fmt.Sscanf(prNumberStr, "%d", &prNumber)
	if err != nil {
		return "", fmt.Errorf("invalid PULL_NUMBER: %s", prNumberStr)
	}

	repo := fmt.Sprintf("%s/%s", owner, repoName)
	return c.GetPRTitle(ctx, repo, prNumber)
}
