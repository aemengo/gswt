package model

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
		if len(suite.FailedTestRuns()) != 0 {
			ts = append(ts, suite)
		}
	}

	return ts
}
