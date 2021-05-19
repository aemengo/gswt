package controller

import (
	"github.com/aemengo/gswt/component"
	"github.com/aemengo/gswt/service"
	"github.com/aemengo/gswt/utils"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
)

type Controller struct {
	svc        *service.Service
	app        *tview.Application
	checksView *component.ChecksView
	logsView   *component.LogsView
}

func New(svc *service.Service, app *tview.Application) *Controller {
	return &Controller{
		svc:        svc,
		app:        app,
		checksView: component.NewChecksView(),
		logsView:   component.NewLogsView(),
	}
}

func (c *Controller) Run() error {
	commits, err := c.svc.Commits()
	if err != nil {
		return err
	}

	checkRuns, err := c.svc.CheckRuns()
	if err != nil {
		return err
	}

	err = c.checksView.Load(c.app, component.ModeChooseChecks, commits, checkRuns)
	if err != nil {
		return err
	}

	go c.handleEvents(commits, checkRuns)

	return c.app.Run()
}

func (c *Controller) handleEvents(commits []*github.RepositoryCommit, checkRuns *github.ListCheckRunsResults) {
	//TODO: handle these errors better

	var (
		chkSuite component.CheckSuite
		logsPath string
	)

	for {
		select {
		case chkSuite = <-c.checksView.CheckSuiteChan:
			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logsPath, err = c.svc.Logs(chkSuite.Selected)
				if err != nil {
					continue
				}
			}

			c.logsView.Load(c.app, component.ModeParseLogs, chkSuite, logsPath)
		case chkSuite = <-c.logsView.LogsCheckSuiteChan:
			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logsPath, err = c.svc.Logs(chkSuite.Selected)
				if err != nil {
					continue
				}
			}

			c.logsView.Load(c.app, component.ModeParseLogs, chkSuite, logsPath)
		case sha := <-c.checksView.SelectedCommitChan:
			var err error
			checkRuns, err = c.svc.CheckRuns(sha)
			if err != nil {
				continue
			}

			c.checksView.Load(c.app, component.ModeChooseChecks, commits, checkRuns)

		// when ESC is pressed
		case <-c.logsView.EscapeLogsTextViewChan:
			c.logsView.Load(c.app, component.ModeChooseChecks, chkSuite, logsPath)
		case <-c.checksView.EscapeCheckListChan:
			c.checksView.Load(c.app, component.ModeChooseCommits, commits, checkRuns)
		case <-c.logsView.EscapeLogsView:
			c.checksView.Load(c.app, component.ModeChooseChecks, commits, checkRuns)
		}

		c.app.Draw()
	}
}
