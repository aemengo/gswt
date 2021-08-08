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
	spec.Run(t, "Parser (stdin)", testParserStdin, spec.Report(report.Terminal{}))
}

func testParser(t *testing.T, _ spec.G, it spec.S) {
	it("parses correctly", func() {
		var (
			id   = 1
			step = model.Step{}
		)

		f, err := os.Open("./parser_test_fixture.txt")
		assertNoError(t, err)
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			step.Lines = append(step.Lines, scanner.Text())
		}

		assertNoError(t, scanner.Err())

		parser := model.NewParser(nil, nil, nil)
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

func testParserStdin(t *testing.T, _ spec.G, it spec.S) {
	it("parses correctly", func() {
		var (
			testSuiteChan = make(chan model.TestSuite, 1)
			doneChan      = make(chan bool, 1)

			collectTestSuites = func() []model.TestSuite {
				var suites []model.TestSuite

				for {
					select {
					case suite := <-testSuiteChan:
						suites = append(suites, suite)
					case <-doneChan:
						return suites
					}
				}
			}
		)

		f, err := os.Open("./parser_test_fixture.txt")
		assertNoError(t, err)
		defer f.Close()

		parser := model.NewParser(testSuiteChan, nil, doneChan)
		go parser.ParseGoTestStdin(f)

		testSuites := collectTestSuites()

		assertNum(t, len(testSuites), 2)

		assertString(t, testSuites[0].Title, "Suite: acceptance-analyzer/0.3 (Passed: 24 | Failed: 2 | Skipped: 7)")
		assertString(t, testSuites[1].Title, "Suite: acceptance-analyzer/0.4 (Passed: 24 | Failed: 2 | Skipped: 7)")

		assertNum(t, len(testSuites[0].TestRuns), 35)
		assertNum(t, len(testSuites[0].FailedTestRuns()), 4)

		assertNum(t, len(testSuites[1].TestRuns), 35)
		assertNum(t, len(testSuites[1].FailedTestRuns()), 4)

		assertNum(t, len(testSuites[0].TestRuns[0].Lines), 5)
		assertNum(t, len(testSuites[1].TestRuns[0].Lines), 5)

		assertNum(t, testSuites[0].TestCount, 33)
		assertNum(t, testSuites[1].TestCount, 33)
	})
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("\nactual: %v\nexpected: nil", err)
	}
}

func assertNum(t *testing.T, actual, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("\nactual: %v\nexpected: %v", actual, expected)
	}
}

func assertString(t *testing.T, actual, expected string) {
	t.Helper()
	if actual != expected {
		t.Errorf("\nactual: %v\nexpected: %v", actual, expected)
	}
}
