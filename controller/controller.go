package controller

import (
	"github.com/aemengo/gswt/component"
	"github.com/aemengo/gswt/service"
	"github.com/aemengo/gswt/utils"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"log"
)

type Controller struct {
	svc        *service.Service
	app        *tview.Application
	checksView *component.ChecksView
	logsView   *component.LogsView
	logger     *log.Logger
}

func New(svc *service.Service, app *tview.Application, logger *log.Logger) *Controller {
	return &Controller{
		svc:        svc,
		app:        app,
		logger:     logger,
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

	c.checksView.Load(c.app, component.ModeChooseChecks, commits, checkRuns)

	go c.handleEvents(commits, checkRuns)

	return c.app.Run()
}

func (c *Controller) handleEvents(commits []*github.RepositoryCommit, checkRuns *github.ListCheckRunsResults) {
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
					c.logger.Println(err)
					continue
				}
			}

			c.logsView.Load(c.app, component.ModeParseLogs, chkSuite, logsPath)
		case chkSuite = <-c.logsView.LogsCheckSuiteChan:
			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logsPath, err = c.svc.Logs(chkSuite.Selected)
				if err != nil {
					c.logger.Println(err)
					continue
				}
			}

			c.logsView.Load(c.app, component.ModeParseLogs, chkSuite, logsPath)
		case sha := <-c.checksView.SelectedCommitChan:
			var err error
			checkRuns, err = c.svc.CheckRuns(sha)
			if err != nil {
				c.logger.Println(err)
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
