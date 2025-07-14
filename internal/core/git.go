package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetGitDirectory finds the git repository root directory by looking for a .git directory
// in the current path or any parent directory. Returns the git root path and true if found,
// or empty string and false if not found.
func GetGitDirectory(path string) (string, bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false, err
	}
	
	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, true, nil
		}
		
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	
	return "", false, nil
}

type Commit struct {
	Hash      string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
}

type CommitPage struct {
	Commits   []Commit
	PageNum   int
	PerPage   int
	HasMore   bool
	Total     int
}

func GetCommitLogs(repoPath string, perPage, pageNum int) (*CommitPage, error) {
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		repoPath = absPath
	}

	gitRoot, isRepo, err := GetGitDirectory(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory is a git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("directory %s is not a git repository", repoPath)
	}
	
	repoPath = gitRoot

	skip := (pageNum - 1) * perPage
	limit := perPage + 1

	format := "--pretty=format:%H|%an|%ae|%at|%s|%b|||END|||"
	
	cmd := exec.Command("git", "-C", repoPath, "log", fmt.Sprintf("--skip=%d", skip), fmt.Sprintf("--max-count=%d", limit), format)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git log: %w", err)
	}

	commits, err := parseCommits(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse commits: %w", err)
	}

	hasMore := len(commits) > perPage
	if hasMore {
		commits = commits[:perPage]
	}

	total, err := getTotalCommitCount(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get total commit count: %w", err)
	}

	return &CommitPage{
		Commits: commits,
		PageNum: pageNum,
		PerPage: perPage,
		HasMore: hasMore,
		Total:   total,
	}, nil
}

func parseCommits(output string) ([]Commit, error) {
	if strings.TrimSpace(output) == "" {
		return []Commit{}, nil
	}

	parts := strings.Split(output, "|||END|||\n")
	commits := make([]Commit, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fields := strings.SplitN(part, "|", 6)
		if len(fields) < 5 {
			continue
		}

		timestamp, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		body := ""
		if len(fields) > 5 {
			body = strings.TrimSpace(fields[5])
		}

		commit := Commit{
			Hash:    fields[0],
			Author:  fields[1],
			Email:   fields[2],
			Date:    time.Unix(timestamp, 0),
			Subject: fields[4],
			Body:    body,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func reverseCommits(commits []Commit) {
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
}

func getTotalCommitCount(repoPath string) (int, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get commit count: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count, nil
}

func GetCommitChangelist(repoPath, commitHash string) ([]byte, error) {
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		repoPath = absPath
	}

	gitRoot, isRepo, err := GetGitDirectory(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory is a git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("directory %s is not a git repository", repoPath)
	}
	
	repoPath = gitRoot

	cmd := exec.Command("git", "-C", repoPath, "show", "--name-status", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changelist for commit %s: %w", commitHash, err)
	}

	return output, nil
}

// GetCommitDiff returns the full diff for a given commit
func GetCommitDiff(repoPath, commitHash string) ([]byte, error) {
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		repoPath = absPath
	}

	gitRoot, isRepo, err := GetGitDirectory(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory is a git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("directory %s is not a git repository", repoPath)
	}
	
	repoPath = gitRoot

	cmd := exec.Command("git", "-C", repoPath, "show", "--format=", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff for commit %s: %w", commitHash, err)
	}

	return output, nil
}

// EstimateTokenCount provides a rough estimate of token count for text
// Uses the approximation that 1 token ≈ 4 characters for English text
func EstimateTokenCount(text string) int {
	return len(text) / 4
}

// FormatTokenCount formats token count in human-readable format (e.g., 2.3k, 1.5M)
func FormatTokenCount(count int) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	} else if count < 1000000 {
		return fmt.Sprintf("%.1fk", float64(count)/1000)
	} else {
		return fmt.Sprintf("%.1fM", float64(count)/1000000)
	}
}

// Changeset represents a commit's changes with metadata
type Changeset struct {
	CommitHash string
	Author     string
	Date       time.Time
	Subject    string
	Body       string
	Diff       string
	Files      []string
}

// GetChangesForCommit retrieves detailed changeset for a specific commit
func GetChangesForCommit(repoPath, commitHash string) (Changeset, error) {
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return Changeset{}, fmt.Errorf("failed to get absolute path: %w", err)
		}
		repoPath = absPath
	}

	gitRoot, isRepo, err := GetGitDirectory(repoPath)
	if err != nil {
		return Changeset{}, fmt.Errorf("failed to check if directory is a git repository: %w", err)
	}
	if !isRepo {
		return Changeset{}, fmt.Errorf("directory %s is not a git repository", repoPath)
	}
	
	repoPath = gitRoot

	// Get commit metadata
	metaCmd := exec.Command("git", "-C", repoPath, "show", "--format=%an|%at|%s|%b", "--no-patch", commitHash)
	metaOutput, err := metaCmd.Output()
	if err != nil {
		return Changeset{}, fmt.Errorf("failed to get commit metadata for %s: %w", commitHash, err)
	}

	// Parse metadata
	metaParts := strings.SplitN(strings.TrimSpace(string(metaOutput)), "|", 4)
	if len(metaParts) < 3 {
		return Changeset{}, fmt.Errorf("invalid commit metadata format")
	}

	timestamp, err := strconv.ParseInt(metaParts[1], 10, 64)
	if err != nil {
		return Changeset{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Get diff
	diff, err := GetCommitDiff(repoPath, commitHash)
	if err != nil {
		return Changeset{}, fmt.Errorf("failed to get diff: %w", err)
	}

	// Get changed files
	filesCmd := exec.Command("git", "-C", repoPath, "show", "--name-only", "--format=", commitHash)
	filesOutput, err := filesCmd.Output()
	if err != nil {
		return Changeset{}, fmt.Errorf("failed to get changed files: %w", err)
	}

	files := []string{}
	for _, file := range strings.Split(string(filesOutput), "\n") {
		file = strings.TrimSpace(file)
		if file != "" {
			files = append(files, file)
		}
	}

	body := ""
	if len(metaParts) > 3 {
		body = strings.TrimSpace(metaParts[3])
	}

	changeset := Changeset{
		CommitHash: commitHash,
		Author:     metaParts[0],
		Date:       time.Unix(timestamp, 0),
		Subject:    metaParts[2],
		Body:       body,
		Diff:       string(diff),
		Files:      files,
	}

	return changeset, nil
}