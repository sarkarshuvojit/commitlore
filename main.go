package main

import (
	"fmt"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/tui"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	
	_, isGitRepo, err := core.GetGitDirectory(cwd)
	if err != nil {
		fmt.Printf("Error checking Git repository: %v\n", err)
		os.Exit(1)
	}
	
	if !isGitRepo {
		fmt.Println("Error: Current directory is not a Git repository")
		os.Exit(1)
	}
	
	if err := tui.RunApp(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}