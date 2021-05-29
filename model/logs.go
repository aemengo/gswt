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

	Lines      []string
	TestSuites []TestSuite
}

type TestSuite struct {
	ID       int
	Title    string
	Selected bool

	TestRuns []TestRun
}

type TestRun struct {
	ID       int
	Name     string
	Success  bool
	Selected bool

	Lines []string
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
		parseGoTest(&logs[i], &id)
	}

	return logs, nil
}

func parseGoTest(step *Step, id *int) {
	var (
		suiteMatcher  = regexp.MustCompile(`^Suite: .+$`)
		tallyMatcher  = regexp.MustCompile(`^Passed: \d+ | Failed: \d+ | Skipped: \d+$`)
		runMatcher    = regexp.MustCompile(`^=== RUN\s+(\S+)$`)
		actionMatcher = regexp.MustCompile(`^=== [A-Z]+\s+(\S+)$`)
		reportMatcher = regexp.MustCompile(`^--- [A-Z]+: (\S+) \(.+$`)
		failedMatcher = regexp.MustCompile(`^\s*--- FAIL: (\S+) \(.+$`)

		runIndexMapping  = map[string]int{}
		currentTestSuite = ""
		currentTestRun   = ""

		// workaround to capture main text before loaded
		mainTestRunName = ""
		mainTestLines   []string
	)

	for _, line := range step.Lines {
		if suiteMatcher.MatchString(line) {
			step.TestSuites = append(step.TestSuites, TestSuite{
				ID:    *id,
				Title: line,

				// placeholder for main TestRun
				TestRuns: []TestRun{
					{
						ID:      *id + 1,
						Name:    mainTestRunName,
						Lines:   mainTestLines,
						Success: true,
					},
				},
			})

			*id = *id + 2

			currentTestSuite = line
			currentTestRun = ""
			continue
		}

		if tallyMatcher.MatchString(line) {
			step.TestSuites[len(step.TestSuites)-1].Title = step.TestSuites[len(step.TestSuites)-1].Title + fmt.Sprintf(" (%s)", line)
			continue
		}

		matches := runMatcher.FindStringSubmatch(line)
		if len(matches) == 2 {
			currentTestRun = matches[1]

			if currentTestSuite == "" {
				mainTestRunName = currentTestRun
				mainTestLines = []string{}
				runIndexMapping[currentTestRun] = 0
				continue
			}

			step.TestSuites[len(step.TestSuites)-1].TestRuns = append(step.TestSuites[len(step.TestSuites)-1].TestRuns, TestRun{
				ID:      *id,
				Name:    currentTestRun,
				Success: true,
			})

			*id = *id + 1

			runIndexMapping[currentTestRun] = len(step.TestSuites[len(step.TestSuites)-1].TestRuns) - 1
			continue
		}

		actionMatches := actionMatcher.FindStringSubmatch(line)
		if len(actionMatches) == 2 {
			if currentTestSuite == "" {
				continue
			}

			currentTestRun = actionMatches[1]
			continue
		}

		if reportMatcher.MatchString(line) {
			currentTestSuite = ""
			currentTestRun = ""
		}

		failureMatches := failedMatcher.FindStringSubmatch(line)
		if len(failureMatches) == 2 {
			testRun := failureMatches[1]

			i, ok := runIndexMapping[testRun]
			if ok {
				if len(step.TestSuites) != 0 {
					step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Success = false
				}
			}

			continue
		}

		if currentTestRun != "" {
			if currentTestSuite == "" {
				mainTestLines = append(mainTestLines, line)
				continue
			}

			i, ok := runIndexMapping[currentTestRun]
			if ok {
				// throw away blank first lines
				testRunLogCount := len(step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines)
				if strings.TrimSpace(line) == "" && testRunLogCount == 0 {
					continue
				}

				step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines = append(step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines, line)
			}
		}
	}
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

func (s *TestSuite) FailedTestRuns() []TestRun {
	var tr []TestRun

	for _, run := range s.TestRuns {
		if !run.Success {
			tr = append(tr, run)
		}
	}
	return tr
}

func bPtr(b bool) *bool {
	return &b
}
