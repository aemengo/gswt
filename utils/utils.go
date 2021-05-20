package utils

import (
	"github.com/google/go-github/v35/github"
	"path/filepath"
)

func LogsDir(homeDir string) string {
	return filepath.Join(homeDir, "logs")
}

func ShouldShowLogs(check *github.CheckRun) bool {
	switch check.GetStatus() {
	case "completed":
		switch check.GetConclusion() {
		case "success", "failure":
			return true
		}
	}

	return false
}
