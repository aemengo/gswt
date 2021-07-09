package controller

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/view"
	"github.com/rivo/tview"
	"io"
	"log"
	"time"
)

type CLController struct {
	app           *tview.Application
	stdin         io.Reader
	testSuiteChan chan model.TestSuite
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

	go model.NewParser(c.testSuiteChan, c.doneChan).ParseGoTestStdin(c.stdin)

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

	for {
		select {
		// when tests are toggled
		case selectedID := <-c.testsView.SelectedStepChan:
			c.logs.Toggle(selectedID)

			selection = view.Selection{Type: view.SelectionTypeID, Value: selectedID}
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)

		// when display mode is toggled
		case m := <-c.testsView.ToggleDisplayModeChan:
			switch m {
			case view.ModeParseTests:
				displayMode = view.ModeParseTestsFuller
			default:
				displayMode = view.ModeParseTests
			}

			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)

		// when user scrolls
		case msg := <-c.testsView.UserDidScrollChan:
			detailText = msg.Msg
			selection = view.Selection{Type: view.SelectionTypeRow, Value: msg.Row}
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)

		// when ticker goes off
		case <-ticker.C:
			c.testsView.UpdateStatus(mode, c.logs, testDuration())

		// when parsing updates
		case testSuite := <-c.testSuiteChan:
			c.logs[0].TestSuites = append(c.logs[0].TestSuites, testSuite)
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)
		case <-c.doneChan:
			mode = view.ModeParseTestsFinished
			ticker.Stop()
			c.endTime = time.Now()
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, selection)
		}

		c.app.Draw()
	}

}
