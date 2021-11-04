package controller

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/utils"
	"github.com/aemengo/gswt/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"log"
	"time"
)

type CLController struct {
	app           *tview.Application
	stdin         io.Reader
	testSuiteChan chan model.TestSuite
	lineChan      chan string
	doneChan      chan bool
	testsView     *view.Tests
	logger        *log.Logger
	logs          model.Logs

	startTime time.Time
	endTime   time.Time
}

func NewCLController(app *tview.Application, logger *log.Logger, stdin io.Reader) *CLController {
	return &CLController{
		app:           app,
		logger:        logger,
		stdin:         stdin,
		testSuiteChan: make(chan model.TestSuite, 1),
		lineChan:      make(chan string, 1),
		doneChan:      make(chan bool, 1),
		testsView:     view.NewTests(),
		startTime:     time.Now(),
		logs: model.Logs{
			model.Step{
				Title:    "go test",
				Selected: true,
				Success:  true,
			},
		},
	}
}

func (c *CLController) Run() error {
	c.testsView.Load(c.app, c.logs, view.ModeParseTestsRunning, view.ModeParseTests, time.Now().Sub(c.startTime), "")

	go c.handleEvents()

		go model.NewParser(c.testSuiteChan, c.lineChan, c.doneChan).ParseGoTestStdin(c.stdin)

	return c.app.Run()
}

func (c *CLController) handleEvents() {
	var (
		mode        = view.ModeParseTestsRunning
		displayMode = view.ModeParseTests
		ticker      = time.NewTicker(250 * time.Millisecond)

		detailText string
		selection  view.Selection

		testDuration = func() time.Duration {
			if !c.endTime.Equal(time.Time{}) {
				return c.endTime.Sub(c.startTime)
			} else {
				return time.Now().Sub(c.startTime)
			}
		}
	)

	// HANDLE USER EVENTS
	// these are unique because app.Draw() cannot be called for these
	// otherwise race conditions will happen
	c.testsView.SetHandlers(
		func(key tcell.Key) {
			if key == tcell.KeyTab {
				c.app.Suspend(func() {
					utils.ShowLogsInEditor(c.logs)
				})
			}
		},
		func() {
			switch displayMode {
			case view.ModeParseTests:
				displayMode = view.ModeParseTestsFuller
			default:
				displayMode = view.ModeParseTests
			}

			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)
		},
		func(id int) {
			c.logs.Toggle(id)
			selection = view.Selection{Type: view.SelectionTypeID, Value: id}
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)
		},
		func(txt string, row int) {
			detailText = txt
			selection = view.Selection{Type: view.SelectionTypeRow, Value: row}
			c.testsView.UpdateDetail(detailText)
		})

	// HANDLE AUTOMATIC EVENTS
	for {
		select {

		// when ticker goes off
		case <-ticker.C:
			c.testsView.UpdateStatus(mode, c.logs, testDuration())

		// when parsing updates
		case line := <-c.lineChan:
			c.logs[0].Lines = append(c.logs[0].Lines, line)
		case testSuite := <-c.testSuiteChan:
			c.logs[0].TestSuites = append(c.logs[0].TestSuites, testSuite)
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)

		// when parsing finishes
		case <-c.doneChan:
			mode = view.ModeParseTestsFinished
			ticker.Stop()
			c.endTime = time.Now()
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)
		}

		c.app.Draw()
	}

}
