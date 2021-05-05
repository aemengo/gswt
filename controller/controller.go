package controller

import (
	"github.com/aemengo/gswt/component"
	"github.com/aemengo/gswt/service"
	"github.com/rivo/tview"
	"os"
)

type Controller struct {
	svc        *service.Service
	app        *tview.Application
	checksView *component.ChecksView
}

func New(svc *service.Service, app *tview.Application) *Controller {
	return &Controller{
		svc:        svc,
		app:        app,
		checksView: component.NewChecksView(),
	}
}

func (c *Controller) Start(sigs <-chan os.Signal) error {
	commits, err := c.svc.Commits()
	if err != nil {
		return err
	}

	checkRuns, err := c.svc.CheckRuns()
	if err != nil {
		return err
	}

	err = c.checksView.Load(c.app, commits, checkRuns)
	if err != nil {
		return err
	}

	err = c.app.Run()
	if err != nil {
		return err
	}

	for {
		select {
		case <-sigs:
			c.app.Stop()
			return nil
		}
	}
}
