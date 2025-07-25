package llm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	
	"github.com/sarkarshuvojit/commitlore/internal/core"
)

// ExtractTopics analyzes changesets and extracts relevant topics for content creation
func ExtractTopics(provider LLMProvider, changesets []Changeset) ([]string, error) {
	if len(changesets) == 0 {
		return []string{}, nil
	}
	
	// Build changeset string from the provided changesets
	changesetString := buildChangesetString(changesets)
	
	systemPrompt := `You are an expert at analyzing git commit changes and extracting meaningful topics for content creation. Your task is to analyze the provided changesets and extract 3-5 key topics that would be interesting for technical blog posts, social media content, or developer stories.

Guidelines:
- Focus on technical achievements, patterns, and insights
- Consider the broader impact and learnings from the changes
- Prioritize topics that would resonate with other developers
- Make topics specific enough to be actionable but broad enough to be interesting
- Return ONLY the topic titles, one per line
- No numbering, bullets, or additional formatting
- Do NOT include any introductory text, explanations, or preamble
- Start immediately with the first topic title`

	userPrompt := fmt.Sprintf("Analyze the following git changesets and extract 3-5 key topics for content creation:\n\n%s", changesetString)
	
	ctx := context.Background()
	response, err := provider.GenerateContentWithSystemPrompt(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract topics from LLM: %w", err)
	}
	
	// Parse the response to extract individual topics
	topics := parseTopicsFromResponse(response)
	
	return topics, nil
}

// buildChangesetString converts changesets into a formatted string for LLM analysis
func buildChangesetString(changesets []Changeset) string {
	var buffer bytes.Buffer
	
	for i, changeset := range changesets {
		buffer.WriteString(fmt.Sprintf("=== Commit %d ===\n", i+1))
		buffer.WriteString(fmt.Sprintf("Hash: %s\n", changeset.CommitHash))
		buffer.WriteString(fmt.Sprintf("Author: %s\n", changeset.Author))
		buffer.WriteString(fmt.Sprintf("Date: %s\n", changeset.Date.Format("2006-01-02 15:04:05")))
		buffer.WriteString(fmt.Sprintf("Subject: %s\n", changeset.Subject))
		
		if changeset.Body != "" {
			buffer.WriteString(fmt.Sprintf("Body: %s\n", changeset.Body))
		}
		
		buffer.WriteString(fmt.Sprintf("Files: %v\n", changeset.Files))
		
		if changeset.Diff != "" {
			// Truncate diff if too long to keep within token limits
			diff := changeset.Diff
			if len(diff) > 2000 {
				diff = diff[:2000] + "\n... (truncated)"
			}
			buffer.WriteString(fmt.Sprintf("Diff:\n%s\n", diff))
		}
		
		buffer.WriteString("\n")
	}
	
	return buffer.String()
}

// parseTopicsFromResponse extracts individual topics from the LLM response
func parseTopicsFromResponse(response string) []string {
	rawLines := strings.Split(response, "\n")
	
	var topics []string
	var skippedLines int
	var shortLines int
	var introductoryLines int
	
	// Common introductory phrases that should be filtered out
	introductoryPhrases := []string{
		"here are",
		"below are", 
		"the following are",
		"i've identified",
		"based on",
		"analyzing",
		"from the",
		"these are",
		"key topics",
		"meaningful topics",
		"extracted topics",
		"relevant topics",
		"topics for",
		"content creation",
		"technical topics",
	}
	
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			skippedLines++
			continue
		}
		
		// Check if this line is likely an introductory phrase
		lowerLine := strings.ToLower(line)
		isIntroductory := false
		for _, phrase := range introductoryPhrases {
			if strings.Contains(lowerLine, phrase) {
				isIntroductory = true
				introductoryLines++
				break
			}
		}
		
		if isIntroductory {
			continue
		}
		
		// Remove common prefixes like numbers, bullets, dashes
		line = strings.TrimLeft(line, "0123456789.-• ")
		line = strings.TrimSpace(line)
		
		// Additional cleanup: remove colons at the end
		line = strings.TrimRight(line, ":")
		line = strings.TrimSpace(line)
		
		if line != "" && len(line) > 10 { // Filter out very short lines
			topics = append(topics, line)
		} else {
			shortLines++
		}
	}
	
	logger := core.GetLogger()
	logger.Debug("Completed topic parsing", 
		"parsed_topics", len(topics),
		"skipped_empty_lines", skippedLines,
		"skipped_short_lines", shortLines,
		"skipped_introductory_lines", introductoryLines)
	
	return topics
}