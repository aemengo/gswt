package utils

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/google/go-github/v35/github"
	"io/ioutil"
	"os"
	"os/exec"
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

func ShowLogsInEditor(logs model.Logs) error {
	f, err := ioutil.TempFile("", "gswt.editor.")
	if err != nil {
		return err
	}

	for _, step := range logs {
		for _, line := range step.Lines {
			f.WriteString(line + "\n")
		}
	}

	f.Close()

	return ShowFileInEditor(f.Name())
}

func ShowFileInEditor(path string) error {
	binaryPath, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		return fmt.Errorf("unable to find path to $EDITOR: %s", err)
	}

	command := exec.Command(binaryPath, path)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	return command.Run()
}