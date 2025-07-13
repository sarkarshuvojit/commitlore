package core

import (
	"os"
	"path/filepath"
)

// IsGitRepository checks if the given path is within a Git repository
// by looking for a .git directory in the current path or any parent directory.
func IsGitRepository(path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	
	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return true, nil
		}
		
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	
	return false, nil
}