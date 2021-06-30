package model

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Parser struct {
	suiteMatcher  *regexp.Regexp
	tallyMatcher  *regexp.Regexp
	runMatcher    *regexp.Regexp
	actionMatcher *regexp.Regexp
	reportMatcher *regexp.Regexp
	failedMatcher *regexp.Regexp

	suiteIndexMapping map[string]int
	runindexMapping   map[string]map[string]int
	currentTestSuite  string
	currentTestRun    string

	// workaround to capture main text before loaded
	mainTestRunName string
	mainTestLines   []string

	doneChan        chan bool
	stepUpdatedChan chan bool
}

func NewParser(stepUpdatedChan chan bool, doneChan chan bool) *Parser {
	return &Parser{
		suiteIndexMapping: map[string]int{},
		runindexMapping:   map[string]map[string]int{},
		stepUpdatedChan:   stepUpdatedChan,
		doneChan:          doneChan,

		suiteMatcher:  regexp.MustCompile(`^Suite: .+$`),
		tallyMatcher:  regexp.MustCompile(`^Passed: \d+ | Failed: \d+ | Skipped: \d+$`),
		runMatcher:    regexp.MustCompile(`^=== RUN\s+(\S+)$`),
		actionMatcher: regexp.MustCompile(`^=== [A-Z]+\s+(\S+)$`),
		reportMatcher: regexp.MustCompile(`^--- [A-Z]+: (\S+) \(.+$`),
		failedMatcher: regexp.MustCompile(`^\s*--- FAIL: (\S+) \(.+$`),
	}
}

func (p *Parser) ParseGoTestStep(id *int, step *Step) {
	for _, line := range step.Lines {
		p.parseGoTestLine(id, step, line)
	}
}

func (p *Parser) ParseGoTestStdin(id *int, step *Step, stdIn io.Reader) {
	scanner := bufio.NewScanner(stdIn)

	for scanner.Scan() {
		p.parseGoTestLine(id, step, scanner.Text())
	}

	// ignore the Err() on purpose
	_ = scanner.Err()
	if p.doneChan != nil {
		p.doneChan <- true
	}
}

// TODO: write tests!
func (p *Parser) parseGoTestLine(id *int, step *Step, line string) {
	if p.suiteMatcher.MatchString(line) {
		step.TestSuites = append(step.TestSuites, TestSuite{
			ID:    *id,
			Title: line,

			TestRuns: []TestRun{
				{
					ID:      *id + 1,
					Name:    p.mainTestRunName,
					Lines:   p.mainTestLines,
					Success: true,
				},
			},
		})

		*id = *id + 2

		p.currentTestSuite = line
		p.currentTestRun = ""
		p.suiteIndexMapping[line] = len(step.TestSuites) - 1

		if p.runindexMapping[line] == nil {
			p.runindexMapping[line] = map[string]int{
				p.mainTestRunName: 0,
			}
		}

		if p.stepUpdatedChan != nil {
			p.stepUpdatedChan <- true
		}

		return
	}

	if p.tallyMatcher.MatchString(line) {
		si, ok := p.suiteIndexMapping[p.currentTestSuite]
		if ok {
			step.TestSuites[si].Title = step.TestSuites[si].Title + fmt.Sprintf(" (%s)", line)
		}

		return
	}

	matches := p.runMatcher.FindStringSubmatch(line)
	if len(matches) == 2 {
		p.currentTestRun = matches[1]

		if p.currentTestSuite == "" {
			p.mainTestRunName = p.currentTestRun
			p.mainTestLines = []string{}
			return
		}

		si, ok := p.suiteIndexMapping[p.currentTestSuite]
		if ok {
			step.TestSuites[si].TestRuns = append(step.TestSuites[si].TestRuns, TestRun{
				ID:      *id,
				Name:    p.currentTestRun,
				Success: true,
			})

			p.runindexMapping[p.currentTestSuite][p.currentTestRun] = len(step.TestSuites[si].TestRuns) - 1

			*id = *id + 1
		}

		return
	}

	actionMatches := p.actionMatcher.FindStringSubmatch(line)
	if len(actionMatches) == 2 {
		if p.currentTestSuite == "" {
			return
		}

		p.currentTestRun = actionMatches[1]
		return
	}

	if p.reportMatcher.MatchString(line) {
		p.currentTestSuite = ""
		p.currentTestRun = ""
	}

	failureMatches := p.failedMatcher.FindStringSubmatch(line)
	if len(failureMatches) == 2 {
		testRun := failureMatches[1]

		for k := range p.runindexMapping {
			ri, ok := p.runindexMapping[k][testRun]
			if ok {
				si := p.suiteIndexMapping[k]

				step.TestSuites[si].TestRuns[ri].Success = false
				return
			}
		}

		return
	}

	if p.currentTestRun != "" {
		si, ok := p.suiteIndexMapping[p.currentTestSuite]
		if !ok {
			p.mainTestLines = append(p.mainTestLines, line)
			return
		}

		ri, ok := p.runindexMapping[p.currentTestSuite][p.currentTestRun]
		if ok {
			// throw away blank first lines
			testRunLogCount := len(step.TestSuites[si].TestRuns[ri].Lines)
			if strings.TrimSpace(line) == "" && testRunLogCount == 0 {
				return
			}

			step.TestSuites[si].TestRuns[ri].Lines = append(step.TestSuites[si].TestRuns[ri].Lines, line)
		}
	}
}
