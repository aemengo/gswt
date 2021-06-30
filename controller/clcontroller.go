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
}

func NewCLController(app *tview.Application, logger *log.Logger, stdin io.Reader) *CLController {
	return &CLController{
		app:             app,
		logger:          logger,
		stdin:           stdin,
		stepUpdatedChan: make(chan bool, 1),
		doneChan:        make(chan bool, 1),
		testsView:       view.NewTests(),
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

	c.testsView.Load(c.app, c.logs, view.ModeParseTestsRunning)

	go c.handleEvents()

	go model.NewParser(c.stepUpdatedChan, c.doneChan).ParseGoTestStdin(&id, &c.logs[0], c.stdin)

	return c.app.Run()
}

func (c *CLController) handleEvents() {
	var (
		mode   = view.ModeParseTestsRunning
		ticker = time.NewTicker(250 * time.Millisecond)
	)

	for {
		select {
		// when tests are toggled
		case id := <-c.testsView.SelectedStepChan:
			c.logs.Toggle(id)
			c.testsView.Load(c.app, c.logs, mode)
		// when ticker goes off
		case <-ticker.C:
			c.testsView.Load(c.app, c.logs, mode)

		// when parsing updates
		case <-c.stepUpdatedChan:
			c.testsView.Load(c.app, c.logs, mode)
		case <-c.doneChan:
			mode = view.ModeParseTestsFinished
			ticker.Stop()
			c.testsView.Load(c.app, c.logs, mode)
		}

		c.app.Draw()
	}

}
