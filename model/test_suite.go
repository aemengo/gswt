package model

type TestSuite struct {
	ID       int
	Title    string
	Selected bool

	TestRuns []TestRun
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