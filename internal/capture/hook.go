package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const postCommitHook = `#!/bin/sh
# vibelog post-commit hook
SESSION_FILE=".vibelog/session.id"
if [ -f "$SESSION_FILE" ]; then
    SESSION_ID=$(cat "$SESSION_FILE")
    vibelog capture --type=file_change --session="$SESSION_ID" --git-diff
fi
`

const postCheckoutHook = `#!/bin/sh
# vibelog post-checkout hook
PREV_HEAD=$1
NEW_HEAD=$2
BRANCH_SWITCH=$3

if [ "$BRANCH_SWITCH" = "1" ]; then
    SESSION_FILE=".vibelog/session.id"
    if [ -f "$SESSION_FILE" ]; then
        SESSION_ID=$(cat "$SESSION_FILE")
        vibelog capture --type=decision --session="$SESSION_ID" --content="Switched branch to $(git branch --show-current)"
    fi
fi
`

func InstallHooks(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	if err := writeHook(hooksDir, "post-commit", postCommitHook); err != nil {
		return fmt.Errorf("post-commit hook: %w", err)
	}

	if err := writeHook(hooksDir, "post-checkout", postCheckoutHook); err != nil {
		return fmt.Errorf("post-checkout hook: %w", err)
	}

	vibelogDir := filepath.Join(repoPath, ".vibelog")
	if err := os.MkdirAll(vibelogDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(vibelogDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := `{"auto_capture": true, "max_events_per_session": 10000}
`
		if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
			return err
		}
	}

	return nil
}

func writeHook(hooksDir, name, content string) error {
	path := filepath.Join(hooksDir, name)

	if existing, err := os.ReadFile(path); err == nil {
		if strings.Contains(string(existing), "vibelog") {
			return nil
		}
		content = string(existing) + "
" + content
	}

	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	return nil
}

func RemoveHooks(repoPath string) error {
	hooksDir := filepath.Join(repoPath, ".git", "hooks")
	for _, name := range []string{"post-commit", "post-checkout"} {
		path := filepath.Join(hooksDir, name)
		if existing, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(existing), "
")
			var filtered []string
			skip := false
			for _, line := range lines {
				if strings.Contains(line, "vibelog") {
					skip = true
					continue
				}
				if skip && strings.TrimSpace(line) == "" {
					skip = false
					continue
				}
				filtered = append(filtered, line)
			}
			if len(filtered) > 0 {
				os.WriteFile(path, []byte(strings.Join(filtered, "
")), 0755)
			} else {
				os.Remove(path)
			}
		}
	}
	return nil
}