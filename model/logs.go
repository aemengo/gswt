package model

import (
	"bufio"
	"os"
	"strings"
)

type Logs []Step

type Step struct {
	ID       int
	Title    string
	Selected bool
	Success  bool
	Lines    []string

	//TODO: do something different about test logs
}

func LogsFromFile(logPath string) (Logs, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		logs           Logs
		shouldCollect  *bool
		id             = 1
		scanner        = bufio.NewScanner(f)
		headerBeginTxt = "##[group]Run "
		headerEndTxt   = "##[endgroup]"
		errorTxt       = "##[error]Process completed with exit code"
		postRunTxt     = "Post job cleanup."
	)

	for scanner.Scan() {
		args := strings.SplitN(scanner.Text(), " ", 2)

		if len(args) != 2 {
			continue
		}

		//timestamp is args[0]
		txt := args[1]

		if strings.HasPrefix(txt, headerBeginTxt) {
			shouldCollect = bPtr(false)

			if len(logs) != 0 {
				if len(logs[len(logs)-1].Lines) == 0 {
					logs[len(logs)-1].Lines = append(logs[len(logs)-1].Lines, "✔︎")
				}
			}

			logs = append(logs, Step{
				ID:      id,
				Title:   strings.TrimPrefix(txt, "##[group]"),
				Success: true,
			})

			id = id + 1
			continue
		}

		if txt == headerEndTxt {
			if shouldCollect != nil && !*shouldCollect {
				shouldCollect = bPtr(true)
			}
			continue
		}

		if txt == postRunTxt {
			shouldCollect = bPtr(false)
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

func (l Logs) Toggle(id int) {
	for i := range l {
		if l[i].ID == id {
			l[i].Selected = !l[i].Selected
		} else {
			l[i].Selected = false
		}
	}
}

func bPtr(b bool) *bool {
	return &b
}
