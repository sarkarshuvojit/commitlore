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