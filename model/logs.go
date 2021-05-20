package model

import (
	"bufio"
	"os"
	"strings"
)

type Logs []Step

type Step struct {
	Title    string
	Selected bool
	Success  bool
	Lines    []string

	//TODO: do something different about test logs
}

func LogsFromFile(logPath string) (Logs, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		logs           Logs
		shouldCollect  *bool
		scanner        = bufio.NewScanner(file)
		headerBeginTxt = "##[group]Run "
		headerEndTxt   = "##[endgroup]"
		errorTxt       = "##[error]Process completed with exit code"
	)

	for scanner.Scan() {
		args := strings.SplitN(scanner.Text(), " ", 2)

		//timestamp is args[0]
		txt := args[1]

		if strings.HasPrefix(txt, headerBeginTxt) {
			shouldCollect = bPtr(false)

			if len(logs) != 0 {
				if len(logs[len(logs)-1].Lines) == 0 {
					logs[len(logs)-1].Lines = append(logs[len(logs)-1].Lines, "-")
				}
			}

			logs = append(logs, Step{
				Title:   strings.TrimPrefix(txt, "##[group]"),
				Success: true,
			})

			continue
		}

		if txt == headerEndTxt {
			if shouldCollect != nil && !*shouldCollect {
				shouldCollect = bPtr(true)
			}
			continue
		}

		if strings.HasPrefix(txt, errorTxt) {
			logs[len(logs)-1].Lines = append(logs[len(logs)-1].Lines, strings.TrimPrefix(txt, "##[error]"))
			logs[len(logs)-1].Success = false
			continue
		}

		if shouldCollect != nil && *shouldCollect {
			logs[len(logs)-1].Lines = append(logs[len(logs)-1].Lines, txt)
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func bPtr(b bool) *bool {
	return &b
}
