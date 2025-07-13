package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIsGitRepository(t *testing.T) {
	t.Run("No Git repository", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "no-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		isGit, err := IsGitRepository(tmpDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if isGit {
			t.Error("Expected false for directory without Git repository")
		}
	})

	t.Run("Git repository in current directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-current-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		isGit, err := IsGitRepository(tmpDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !isGit {
			t.Error("Expected true for directory with .git directory")
		}
	})

	t.Run("Git repository few levels up", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-parent-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create .git in root temp directory
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		// Create nested subdirectories
		subDir1 := filepath.Join(tmpDir, "level1")
		subDir2 := filepath.Join(subDir1, "level2")
		subDir3 := filepath.Join(subDir2, "level3")

		if err := os.MkdirAll(subDir3, 0755); err != nil {
			t.Fatalf("Failed to create nested directories: %v", err)
		}

		// Test from level 1
		isGit, err := IsGitRepository(subDir1)
		if err != nil {
			t.Fatalf("Unexpected error from level1: %v", err)
		}
		if !isGit {
			t.Error("Expected true for level1 subdirectory with Git repository above")
		}

		// Test from level 2
		isGit, err = IsGitRepository(subDir2)
		if err != nil {
			t.Fatalf("Unexpected error from level2: %v", err)
		}
		if !isGit {
			t.Error("Expected true for level2 subdirectory with Git repository above")
		}

		// Test from level 3
		isGit, err = IsGitRepository(subDir3)
		if err != nil {
			t.Fatalf("Unexpected error from level3: %v", err)
		}
		if !isGit {
			t.Error("Expected true for level3 subdirectory with Git repository above")
		}
	})

	t.Run("Git repository in subdirectory but not parent", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-sub-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create subdirectory with .git
		subDir := filepath.Join(tmpDir, "project")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		gitDir := filepath.Join(subDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		// Test from parent directory (should return false)
		isGit, err := IsGitRepository(tmpDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if isGit {
			t.Error("Expected false for parent directory without .git")
		}

		// Test from subdirectory (should return true)
		isGit, err = IsGitRepository(subDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !isGit {
			t.Error("Expected true for subdirectory with .git")
		}
	})

	t.Run("Invalid path", func(t *testing.T) {
		// Test with non-existent path
		isGit, err := IsGitRepository("/non/existent/path")
		if err != nil {
			t.Fatalf("Unexpected error for non-existent path: %v", err)
		}
		if isGit {
			t.Error("Expected false for non-existent path")
		}
	})

	t.Run("Relative path", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "git-relative-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		// Change to temp directory
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(oldWd)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Test with relative path
		isGit, err := IsGitRepository(".")
		if err != nil {
			t.Fatalf("Unexpected error with relative path: %v", err)
		}
		if !isGit {
			t.Error("Expected true for relative path to Git repository")
		}
	})

	t.Run("Empty path", func(t *testing.T) {
		// Test with empty string (should use current directory)
		isGit, err := IsGitRepository("")
		if err != nil {
			t.Fatalf("Unexpected error with empty path: %v", err)
		}
		// Result depends on whether test is run in a Git repository
		// We just verify no error occurs
		_ = isGit
	})
}

func createTestRepo(t *testing.T) string {
	t.Helper()
	
	tmpDir := t.TempDir()

	if err := exec.Command("git", "-C", tmpDir, "init").Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	if err := exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("Failed to set git user name: %v", err)
	}

	if err := exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("Failed to set git user email: %v", err)
	}

	for i := 1; i <= 20; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		content := fmt.Sprintf("This is file %d\nContent for commit %d", i, i)
		
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}

		if err := exec.Command("git", "-C", tmpDir, "add", filename).Run(); err != nil {
			t.Fatalf("Failed to add file %s: %v", filename, err)
		}

		commitMsg := fmt.Sprintf("Commit %d: Add %s", i, filename)
		if err := exec.Command("git", "-C", tmpDir, "commit", "-m", commitMsg).Run(); err != nil {
			t.Fatalf("Failed to commit file %s: %v", filename, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	return tmpDir
}

func TestGetCommitLogs(t *testing.T) {
	repoPath := createTestRepo(t)

	t.Run("First page with 5 commits", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 5, 1)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if page.PageNum != 1 {
			t.Errorf("Expected PageNum 1, got %d", page.PageNum)
		}
		if page.PerPage != 5 {
			t.Errorf("Expected PerPage 5, got %d", page.PerPage)
		}
		if len(page.Commits) != 5 {
			t.Errorf("Expected 5 commits, got %d", len(page.Commits))
		}
		if page.Total != 20 {
			t.Errorf("Expected Total 20, got %d", page.Total)
		}
		if !page.HasMore {
			t.Error("Expected HasMore to be true")
		}

		if page.Commits[0].Subject != "Commit 20: Add file20.txt" {
			t.Errorf("Expected first commit to be 'Commit 20: Add file20.txt', got '%s'", page.Commits[0].Subject)
		}
		if page.Commits[4].Subject != "Commit 16: Add file16.txt" {
			t.Errorf("Expected fifth commit to be 'Commit 16: Add file16.txt', got '%s'", page.Commits[4].Subject)
		}
	})

	t.Run("Second page with 5 commits", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 5, 2)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if page.PageNum != 2 {
			t.Errorf("Expected PageNum 2, got %d", page.PageNum)
		}
		if len(page.Commits) != 5 {
			t.Errorf("Expected 5 commits, got %d", len(page.Commits))
		}
		if !page.HasMore {
			t.Error("Expected HasMore to be true")
		}

		if page.Commits[0].Subject != "Commit 15: Add file15.txt" {
			t.Errorf("Expected first commit to be 'Commit 15: Add file15.txt', got '%s'", page.Commits[0].Subject)
		}
		if page.Commits[4].Subject != "Commit 11: Add file11.txt" {
			t.Errorf("Expected fifth commit to be 'Commit 11: Add file11.txt', got '%s'", page.Commits[4].Subject)
		}
	})

	t.Run("Last page with 5 commits", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 5, 4)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if page.PageNum != 4 {
			t.Errorf("Expected PageNum 4, got %d", page.PageNum)
		}
		if len(page.Commits) != 5 {
			t.Errorf("Expected 5 commits, got %d", len(page.Commits))
		}
		if page.HasMore {
			t.Error("Expected HasMore to be false")
		}

		if page.Commits[0].Subject != "Commit 5: Add file5.txt" {
			t.Errorf("Expected first commit to be 'Commit 5: Add file5.txt', got '%s'", page.Commits[0].Subject)
		}
		if page.Commits[4].Subject != "Commit 1: Add file1.txt" {
			t.Errorf("Expected fifth commit to be 'Commit 1: Add file1.txt', got '%s'", page.Commits[4].Subject)
		}
	})

	t.Run("Page size 10", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 10, 1)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if len(page.Commits) != 10 {
			t.Errorf("Expected 10 commits, got %d", len(page.Commits))
		}
		if !page.HasMore {
			t.Error("Expected HasMore to be true")
		}

		page2, err := GetCommitLogs(repoPath, 10, 2)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}

		if len(page2.Commits) != 10 {
			t.Errorf("Expected 10 commits on page 2, got %d", len(page2.Commits))
		}
		if page2.HasMore {
			t.Error("Expected HasMore to be false on page 2")
		}
	})

	t.Run("Page size 3", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 3, 1)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if len(page.Commits) != 3 {
			t.Errorf("Expected 3 commits, got %d", len(page.Commits))
		}
		if !page.HasMore {
			t.Error("Expected HasMore to be true")
		}

		if page.Commits[0].Subject != "Commit 20: Add file20.txt" {
			t.Errorf("Expected first commit to be 'Commit 20: Add file20.txt', got '%s'", page.Commits[0].Subject)
		}
		if page.Commits[2].Subject != "Commit 18: Add file18.txt" {
			t.Errorf("Expected third commit to be 'Commit 18: Add file18.txt', got '%s'", page.Commits[2].Subject)
		}
	})

	t.Run("Out of bounds page", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 5, 10)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if len(page.Commits) != 0 {
			t.Errorf("Expected 0 commits for out of bounds page, got %d", len(page.Commits))
		}
		if page.HasMore {
			t.Error("Expected HasMore to be false for out of bounds page")
		}
	})

	t.Run("Verify descending order", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 20, 1)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if len(page.Commits) != 20 {
			t.Errorf("Expected 20 commits, got %d", len(page.Commits))
		}

		var prevTime time.Time
		for i, commit := range page.Commits {
			if i == 0 {
				prevTime = commit.Date
				continue
			}

			if commit.Date.After(prevTime) {
				t.Errorf("Commits not in descending order: commit %d (%s) is after commit %d (%s)", 
					i, commit.Date.Format(time.RFC3339), i-1, prevTime.Format(time.RFC3339))
			}
			prevTime = commit.Date
		}

		if page.Commits[0].Subject != "Commit 20: Add file20.txt" {
			t.Errorf("Expected newest commit first, got '%s'", page.Commits[0].Subject)
		}
		if page.Commits[19].Subject != "Commit 1: Add file1.txt" {
			t.Errorf("Expected oldest commit last, got '%s'", page.Commits[19].Subject)
		}
	})

	t.Run("Non-git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := GetCommitLogs(tmpDir, 5, 1)
		if err == nil {
			t.Error("Expected error for non-git repository")
		}
	})

	t.Run("Invalid parameters", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 0, 1)
		if err != nil {
			t.Fatalf("Failed with perPage 0: %v", err)
		}
		if len(page.Commits) != 0 {
			t.Errorf("Expected 0 commits with perPage 0, got %d", len(page.Commits))
		}

		page, err = GetCommitLogs(repoPath, 5, 0)
		if err != nil {
			t.Fatalf("Failed with pageNum 0: %v", err)
		}
		if len(page.Commits) != 5 {
			t.Errorf("Expected 5 commits with pageNum 0, got %d", len(page.Commits))
		}
	})

	t.Run("Verify commit metadata", func(t *testing.T) {
		page, err := GetCommitLogs(repoPath, 1, 1)
		if err != nil {
			t.Fatalf("Failed to get commit logs: %v", err)
		}

		if len(page.Commits) != 1 {
			t.Fatalf("Expected 1 commit, got %d", len(page.Commits))
		}

		commit := page.Commits[0]
		if commit.Hash == "" {
			t.Error("Expected non-empty hash")
		}
		if commit.Author != "Test User" {
			t.Errorf("Expected author 'Test User', got '%s'", commit.Author)
		}
		if commit.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", commit.Email)
		}
		if commit.Date.IsZero() {
			t.Error("Expected non-zero date")
		}
		if commit.Subject == "" {
			t.Error("Expected non-empty subject")
		}
	})
}