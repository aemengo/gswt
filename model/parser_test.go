package model_test

import (
	"bufio"
	"github.com/aemengo/gswt/model"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	spec.Run(t, "Parser", testParser, spec.Report(report.Terminal{}))
}

func testParser(t *testing.T, when spec.G, it spec.S) {
	it("is true", func() {
		var id = 1
		var step = model.Step{}

		f, err := os.Open("./parser_test_fixture.txt")
		assertNoError(t, err)
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			step.Lines = append(step.Lines, scanner.Text())
		}
		assertNoError(t, scanner.Err())

		parser := model.NewParser(nil, nil)
		parser.ParseGoTestStep(&id, &step)

		assertNum(t, len(step.TestSuites), 2)

		assertString(t, step.TestSuites[0].Title, "Suite: acceptance-analyzer/0.3 (Passed: 24 | Failed: 2 | Skipped: 7)")
		assertString(t, step.TestSuites[1].Title, "Suite: acceptance-analyzer/0.4 (Passed: 24 | Failed: 2 | Skipped: 7)")

		assertNum(t, len(step.TestSuites[0].TestRuns), 35)
		assertNum(t, len(step.TestSuites[0].FailedTestRuns()), 4)

		assertNum(t, len(step.TestSuites[1].TestRuns), 35)
		assertNum(t, len(step.TestSuites[1].FailedTestRuns()), 4)

		assertNum(t, len(step.TestSuites[0].TestRuns[0].Lines), 5)
		assertNum(t, len(step.TestSuites[1].TestRuns[0].Lines), 5)

		assertNum(t, step.TestSuites[0].TestCount, 33)
		assertNum(t, step.TestSuites[1].TestCount, 33)
	})
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("\nactual: %v\nexpected: nil", err)
	}
}

func assertNum(t *testing.T, actual, expected int) {
	if actual != expected {
		t.Errorf("\nactual: %v\nexpected: %v", actual, expected)
	}
}

func assertString(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Errorf("\nactual: %v\nexpected: %v", actual, expected)
	}
}
