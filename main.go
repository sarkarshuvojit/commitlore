package main

import (
	"fmt"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/tui"
)

func main() {
	if err := core.InitLogger(); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	
	logger := core.GetLogger()
	logger.Info("CommitLore application starting")
	
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error getting current directory", "error", err)
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	
	_, isGitRepo, err := core.GetGitDirectory(cwd)
	if err != nil {
		logger.Error("Error checking Git repository", "error", err)
		fmt.Printf("Error checking Git repository: %v\n", err)
		os.Exit(1)
	}
	
	if !isGitRepo {
		logger.Error("Current directory is not a Git repository", "path", cwd)
		fmt.Println("Error: Current directory is not a Git repository")
		os.Exit(1)
	}
	
	logger.Info("Starting TUI application", "repository", cwd)
	if err := tui.RunApp(); err != nil {
		logger.Error("TUI application error", "error", err)
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	
	logger.Info("CommitLore application completed successfully")
}