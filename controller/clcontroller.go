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
	app             *tview.Application
	stdin           io.Reader
	stepUpdatedChan chan bool
	doneChan        chan bool
	testsView       *view.Tests
	logger          *log.Logger
	logs            model.Logs

	startTime time.Time
	endTime   time.Time
}

func NewCLController(app *tview.Application, logger *log.Logger, stdin io.Reader) *CLController {
	return &CLController{
		app:             app,
		logger:          logger,
		stdin:           stdin,
		stepUpdatedChan: make(chan bool, 1),
		doneChan:        make(chan bool, 1),
		testsView:       view.NewTests(),
		startTime:       time.Now(),
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
	var id = 1

	c.testsView.Load(c.app, c.logs, view.ModeParseTestsRunning, view.ModeParseTests, time.Now().Sub(c.startTime), "")

	go c.handleEvents()

	go model.NewParser(c.stepUpdatedChan, c.doneChan).ParseGoTestStdin(&id, &c.logs[0], c.stdin)

	return c.app.Run()
}

func (c *CLController) handleEvents() {
	var (
		mode        = view.ModeParseTestsRunning
		displayMode = view.ModeParseTests
		ticker      = time.NewTicker(250 * time.Millisecond)
		detailText  = ""
		selectedID  = 0

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
		case selectedID = <-c.testsView.SelectedStepChan:
			c.logs.Toggle(selectedID)
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, view.Selection{Type: view.SelectionTypeID, Value: selectedID})

		// when display mode is toggled
		case m := <-c.testsView.ToggleDisplayModeChan:
			switch m {
			case view.ModeParseTests:
				displayMode = view.ModeParseTestsFuller
			default:
				displayMode = view.ModeParseTests
				detailText = ""
			}

			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, view.Selection{Type: view.SelectionTypeID, Value: selectedID})

		// when user scrolls
		case msg := <-c.testsView.UserDidScrollChan:
			detailText = msg.Msg
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText, view.Selection{Type: view.SelectionTypeRow, Value: msg.Row})

		// when ticker goes off
		case <-ticker.C:
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText)

		// when parsing updates
		case <-c.stepUpdatedChan:
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText)
		case <-c.doneChan:
			mode = view.ModeParseTestsFinished
			ticker.Stop()
			c.endTime = time.Now()
			c.testsView.Load(c.app, c.logs, mode, displayMode, testDuration(), detailText)
		}

		c.app.Draw()
	}

}
