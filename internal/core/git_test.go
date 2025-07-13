package core

import (
	"os"
	"path/filepath"
	"testing"
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