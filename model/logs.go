package model

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Logs []Step

type Step struct {
	ID       int
	Title    string
	Selected bool
	Success  bool
	Lines    []string

	TestSuites []TestSuite
}

type TestSuite struct {
	ID       int
	Title    string
	Selected bool
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

	for i := range logs {
		parseGoTest(&logs[i], &id)
	}

	return logs, nil
}

func parseGoTest(step *Step, id *int) {
	var (
		//shouldCollect *bool
		suiteMatcher = regexp.MustCompile(`^Suite: .+$`)
		tallyMatcher = regexp.MustCompile(`^Passed: \d+ | Failed: \d+ | Skipped: \d+$`)
	)

	for _, line := range step.Lines {
		if suiteMatcher.MatchString(line) {
			//shouldCollect = bPtr(true)

			step.TestSuites = append(step.TestSuites, TestSuite{
				ID:    *id,
				Title: line,
			})

			*id = *id + 1
			continue
		}

		if tallyMatcher.MatchString(line) {
			step.TestSuites[len(step.TestSuites)-1].Title = step.TestSuites[len(step.TestSuites)-1].Title + fmt.Sprintf(" (%s)", line)
			continue
		}
	}

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

func (s *Step) IsTest() bool {
	return len(s.TestSuites) != 0
}

func (s *Step) FailedTestSuites() []TestSuite {
	var ts []TestSuite

	for _, suite := range s.TestSuites {
		if !strings.Contains(suite.Title, "Failed: 0") {
			ts = append(ts, suite)
		}
	}

	return ts
}

func bPtr(b bool) *bool {
	return &b
}
