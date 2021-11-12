package model

import "strings"

type Step struct {
	ID       int
	Title    string
	Selected bool
	Success  bool

	Lines      []string
	TestSuites []TestSuite
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

func (s *Step) HaveUnhandledFailures() bool {
	return len(s.FailedTestSuites()) == 0 && func() bool {
		for _, line := range s.Lines {
			if strings.Contains(line, "FAIL") {
				return true
			}
		}
		return false
	}()
}
