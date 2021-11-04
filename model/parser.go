package model

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	suiteMatcher  *regexp.Regexp
	tallyMatcher  *regexp.Regexp
	totalMatcher  *regexp.Regexp
	runMatcher    *regexp.Regexp
	actionMatcher *regexp.Regexp
	reportMatcher *regexp.Regexp
	failedMatcher *regexp.Regexp

	suiteIndexMapping map[string]int
	runIndexMapping   map[string]map[string]int
	currentTestSuite  string
	currentTestRun    string

	// workaround to capture main text before loaded
	mainTestRunName string
	mainTestLines   []string

	// workaround to capture panic information
	extraLines []string

	doneChan       chan bool
	testSuiteChan  chan TestSuite
	lineChan       chan string
	testSuiteIndex int
}

func NewParser(testSuiteChan chan TestSuite, lineChan chan string, doneChan chan bool) *Parser {
	return &Parser{
		suiteIndexMapping: map[string]int{},
		runIndexMapping:   map[string]map[string]int{},
		testSuiteChan:     testSuiteChan,
		lineChan:          lineChan,
		doneChan:          doneChan,
		testSuiteIndex:    0,

		suiteMatcher:  regexp.MustCompile(`^Suite: .+$`),
		tallyMatcher:  regexp.MustCompile(`^Passed: \d+ | Failed: \d+ | Skipped: \d+$`),
		totalMatcher:  regexp.MustCompile(`^Total: (\d+) | Focused: \d+ | Pending: \d+$`),
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

func (p *Parser) ParseGoTestStdin(stdin io.Reader) {
	var (
		id      = 1
		scanner = bufio.NewScanner(stdin)
		step    = Step{}
	)

	for scanner.Scan() {
		p.parseGoTestLine(&id, &step, scanner.Text())
	}

	f, _ := os.Create("/tmp/anthony.txt")
	for _, line := range step.ExtraLines {
		f.WriteString(fmt.Sprintf("[%s]\n", line))
	}

	//fmt.Println("BANANAAD", len(step.ExtraLines))

	// ignore the Err() on purpose
	// _ = scanner.Err()
	if p.doneChan != nil {
		p.sendTestSuites(&step)

		time.Sleep(time.Millisecond)
		p.doneChan <- true
	}
}

func (p *Parser) parseGoTestLine(id *int, step *Step, line string) {
	defer p.sendLine(line)

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

		if p.runIndexMapping[line] == nil {
			p.runIndexMapping[line] = map[string]int{
				p.mainTestRunName: 0,
			}
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

	totalMatches := p.totalMatcher.FindStringSubmatch(line)
	if len(totalMatches) == 2 {
		totalStr := totalMatches[1]

		si, ok := p.suiteIndexMapping[p.currentTestSuite]
		if ok {
			total, _ := strconv.Atoi(totalStr)
			step.TestSuites[si].TestCount = step.TestSuites[si].TestCount + total
		}

		return
	}

	runMatches := p.runMatcher.FindStringSubmatch(line)
	if len(runMatches) == 2 {
		p.currentTestRun = runMatches[1]

		if p.currentTestSuite == "" {
			// This represents parsing a new set of test suites
			p.mainTestRunName = p.currentTestRun
			p.mainTestLines = []string{}
			p.sendTestSuites(step)
			return
		}

		si, ok := p.suiteIndexMapping[p.currentTestSuite]
		if ok {
			step.TestSuites[si].TestRuns = append(step.TestSuites[si].TestRuns, TestRun{
				ID:      *id,
				Name:    p.currentTestRun,
				Success: true,
			})

			p.runIndexMapping[p.currentTestSuite][p.currentTestRun] = len(step.TestSuites[si].TestRuns) - 1

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

		for k, si := range p.suiteIndexMapping {
			ri, ok := p.runIndexMapping[k][testRun]
			if ok {
				step.TestSuites[si].TestRuns[ri].Success = false
			}
		}

		return
	}

	switch {
	case p.currentTestSuite == "" && p.currentTestRun == "":
		//p.extraLines = append(p.extraLines, line)
		step.ExtraLines = append(step.ExtraLines, line)

	case p.currentTestSuite == "" && p.currentTestRun != "":
		p.mainTestLines = append(p.mainTestLines, line)

	case p.currentTestRun != "":
		for k, si := range p.suiteIndexMapping {
			ri, ok := p.runIndexMapping[k][p.currentTestRun]
			if ok {
				// throw away blank first lines
				testRunLogCount := len(step.TestSuites[si].TestRuns[ri].Lines)
				if strings.TrimSpace(line) == "" && testRunLogCount == 0 {
					continue
				}

				step.TestSuites[si].TestRuns[ri].Lines = append(step.TestSuites[si].TestRuns[ri].Lines, line)
			}
		}
	}
}

func (p *Parser) sendTestSuites(step *Step) {
	var i int

	for i = p.testSuiteIndex; i <= len(step.TestSuites)-1; i++ {
		if p.testSuiteChan != nil {
			p.testSuiteChan <- step.TestSuites[i]
		}
	}

	p.testSuiteIndex = i
}

func (p *Parser) sendLine(line string) {
	if p.lineChan != nil {
		p.lineChan <- line
	}
}