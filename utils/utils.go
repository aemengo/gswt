package utils

import "github.com/google/go-github/v35/github"

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
