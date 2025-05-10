package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func PromptForSessionName(sessionNames map[string]bool) (string, error) {
	if len(sessionNames) == 0 {
		return "", fmt.Errorf("no session names to select from")
	}
	uniqueNames := []string{}

	for key := range sessionNames {
		uniqueNames = append(uniqueNames, key)
	}
	defaultSelect := uniqueNames[0]
	inputString := strings.Join(uniqueNames, "\n")

	// fmt.Println(inputString, defaultSelect)

	cmd := exec.Command("fzf",
		"--tmux", "80%",
		// "--layout=reverse",
		"--prompt", "Session name: ",
		"--header", "Enter to select, type to create new",
		"--print-query",            // Print user's query if nothing matches
		// "--select-1",               // Auto-select if only one match
		"--query", defaultSelect) // Pre-populate with suggestion

	cmd.Stdin = strings.NewReader(inputString)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("session name selection canceled: %v", err)
		}
		return "", fmt.Errorf("failed to run fzf: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no session names selected")
	}

	sessionName := lines[0]

	if sessionName == "" {
		return "", fmt.Errorf("session name cannot be empty")
	}

	sessionName = sanitizeSessionName(sessionName)

	return sessionName, nil
}

func sanitizeSessionName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Replace colons and spaces with underscores
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, " ", "_")

	// Remove other problematic characters
	// Add more replacements if needed for other invalid characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	return name
}
