package controller

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/service"
	"github.com/aemengo/gswt/utils"
	"github.com/aemengo/gswt/view"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"log"
)

type Controller struct {
	svc        *service.Service
	app        *tview.Application
	checksView *view.Checks
	logsView   *view.Logs
	logger     *log.Logger
}

func New(svc *service.Service, app *tview.Application, logger *log.Logger) *Controller {
	return &Controller{
		svc:        svc,
		app:        app,
		logger:     logger,
		checksView: view.NewChecks(svc),
		logsView:   view.NewLogs(),
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

	c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns)

	go c.handleEvents(commits, checkRuns)

	return c.app.Run()
}

func (c *Controller) handleEvents(commits []*github.RepositoryCommit, checkRuns *github.ListCheckRunsResults) {
	var (
		chkSuite  model.CheckSuite
		commitSHA string

		logs    model.Logs
		logMode = view.ModeParseLogs

		detailText string
		selection  view.Selection
	)

	for {
		select {
		// AUTOMATIC EVENTS
		// when workflows are done loading
		case <-c.svc.FetchChan:
			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns)

		// USER EVENTS
		// when commits are picked
		case commitSHA = <-c.checksView.SelectedCommitChan:
			var err error
			checkRuns, err = c.svc.CheckRuns(commitSHA)
			if err != nil {
				c.logger.Println(err)
				continue
			}

			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns, commitSHA)
		// when checks are picked
		case chkSuite = <-c.checksView.CheckSuiteChan:
			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logs, err = c.fetchLogs(chkSuite.Selected)
				if err != nil {
					c.logger.Println(err)
					continue
				}
			}

			c.logsView.Load(c.app, view.ModeParseLogs, chkSuite, logs, detailText)
		case chkSuite = <-c.logsView.LogsCheckSuiteChan:
			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logs, err = c.fetchLogs(chkSuite.Selected)
				if err != nil {
					c.logger.Println(err)
					continue
				}
			}

			c.logsView.Load(c.app, view.ModeParseLogs, chkSuite, logs, detailText)
		// when logs are toggled
		case selectedID := <-c.logsView.SelectedStepChan:
			logs.Toggle(selectedID)

			selection = view.Selection{Type: view.SelectionTypeID, Value: selectedID}
			c.logsView.Load(c.app, logMode, chkSuite, logs, detailText, selection)
		case m := <-c.logsView.ToggleModeChan:
			switch m {
			case view.ModeParseLogs:
				logMode = view.ModeParseLogsFuller
			default:
				logMode = view.ModeParseLogs
			}

			c.logsView.Load(c.app, logMode, chkSuite, logs, detailText, selection)
		// when user scrolls
		case msg := <-c.logsView.UserDidScrollChan:
			detailText = msg.Msg
			selection = view.Selection{Type: view.SelectionTypeRow, Value: msg.Row}
			c.logsView.Load(c.app, logMode, chkSuite, logs, detailText, selection)
		// when ESC is pressed
		case <-c.logsView.EscapeLogsDetailChan:
			logMode = view.ModeParseLogs
			c.logsView.Load(c.app, view.ModeChooseChecks, chkSuite, logs, detailText)
		case <-c.checksView.EscapeCheckListChan:
			c.checksView.Load(c.app, view.ModeChooseCommits, commits, checkRuns, commitSHA)
		case <-c.logsView.EscapeLogsChan:
			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns, commitSHA)
		}

		c.app.Draw()
	}
}

func (c *Controller) fetchLogs(checkRun *github.CheckRun) (model.Logs, error) {
	logsPath, err := c.svc.Logs(checkRun)
	if err != nil {
		return nil, err
	}

	return model.LogsFromFile(logsPath)
}
