package model

import (
	"fmt"
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

	id               *int
	runIndexMapping  map[string]int
	currentTestSuite string
	currentTestRun   string

	// workaround to capture main text before loaded
	mainTestRunName string
	mainTestLines   []string
}

func NewParser(id *int) *Parser {
	return &Parser{
		id:              id,
		runIndexMapping: map[string]int{},
		suiteMatcher:    regexp.MustCompile(`^Suite: .+$`),
		tallyMatcher:    regexp.MustCompile(`^Passed: \d+ | Failed: \d+ | Skipped: \d+$`),
		runMatcher:      regexp.MustCompile(`^=== RUN\s+(\S+)$`),
		actionMatcher:   regexp.MustCompile(`^=== [A-Z]+\s+(\S+)$`),
		reportMatcher:   regexp.MustCompile(`^--- [A-Z]+: (\S+) \(.+$`),
		failedMatcher:   regexp.MustCompile(`^\s*--- FAIL: (\S+) \(.+$`),
	}
}

func (p *Parser) ParseGoTestStep(step *Step) {
	for _, line := range step.Lines {
		p.parseGoTestLine(step, line)
	}
}

func (p *Parser) parseGoTestLine(step *Step, line string) {
	if p.suiteMatcher.MatchString(line) {
		step.TestSuites = append(step.TestSuites, TestSuite{
			ID:    *p.id,
			Title: line,

			// placeholder for main TestRun
			TestRuns: []TestRun{
				{
					ID:      *p.id + 1,
					Name:    p.mainTestRunName,
					Lines:   p.mainTestLines,
					Success: true,
				},
			},
		})

		*p.id = *p.id + 2

		p.currentTestSuite = line
		p.currentTestRun = ""
		return
	}

	if p.tallyMatcher.MatchString(line) {
		step.TestSuites[len(step.TestSuites)-1].Title = step.TestSuites[len(step.TestSuites)-1].Title + fmt.Sprintf(" (%s)", line)
		return
	}

	matches := p.runMatcher.FindStringSubmatch(line)
	if len(matches) == 2 {
		p.currentTestRun = matches[1]

		if p.currentTestSuite == "" {
			p.mainTestRunName = p.currentTestRun
			p.mainTestLines = []string{}
			p.runIndexMapping[p.currentTestRun] = 0
			return
		}

		step.TestSuites[len(step.TestSuites)-1].TestRuns = append(step.TestSuites[len(step.TestSuites)-1].TestRuns, TestRun{
			ID:      *p.id,
			Name:    p.currentTestRun,
			Success: true,
		})

		*p.id = *p.id + 1

		p.runIndexMapping[p.currentTestRun] = len(step.TestSuites[len(step.TestSuites)-1].TestRuns) - 1
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

		i, ok := p.runIndexMapping[testRun]
		if ok {
			if len(step.TestSuites) != 0 {
				step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Success = false
			}
		}

		return
	}

	if p.currentTestRun != "" {
		if p.currentTestSuite == "" {
			p.mainTestLines = append(p.mainTestLines, line)
			return
		}

		i, ok := p.runIndexMapping[p.currentTestRun]
		if ok {
			// throw away blank first lines
			testRunLogCount := len(step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines)
			if strings.TrimSpace(line) == "" && testRunLogCount == 0 {
				return
			}

			step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines = append(step.TestSuites[len(step.TestSuites)-1].TestRuns[i].Lines, line)
		}
	}
}
