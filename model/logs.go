package model

import (
	"bufio"
	"os"
	"strings"
)

type Logs []Step

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

			logs = append(logs, Step{
				ID:      id,
				Title:   strings.TrimPrefix(txt, headerBeginTxt),
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

	for i := range logs {
		NewParser(nil, nil).ParseGoTestStep(&id, &logs[i])
	}

	return logs, nil
}

func (l Logs) toggleTestRuns(id int) bool {
	var found bool

	for i := range l {
		for j := range l[i].TestSuites {
			for k := range l[i].TestSuites[j].TestRuns {
				if l[i].TestSuites[j].TestRuns[k].ID == id {
					l[i].TestSuites[j].TestRuns[k].Selected = !l[i].TestSuites[j].TestRuns[k].Selected
					found = true
				} else {
					l[i].TestSuites[j].TestRuns[k].Selected = false
				}
			}
		}
	}

	return found
}

func (l Logs) toggleTestSuites(id int) bool {
	var found bool

	for i := range l {
		for j := range l[i].TestSuites {
			if l[i].TestSuites[j].ID == id {
				l[i].TestSuites[j].Selected = !l[i].TestSuites[j].Selected
				found = true
			} else {
				l[i].TestSuites[j].Selected = false
			}
		}
	}

	return found
}

func (l Logs) Toggle(id int) {
	ok := l.toggleTestRuns(id)
	if ok {
		return
	}

	ok = l.toggleTestSuites(id)
	if ok {
		return
	}

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
