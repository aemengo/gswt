package controller

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/service"
	"github.com/aemengo/gswt/utils"
	"github.com/aemengo/gswt/view"
	"github.com/gdamore/tcell/v2"
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

	// HANDLE USER EVENTS
	// these are unique because app.Draw() cannot be called for these
	// otherwise race conditions will happen
	// checksView
	c.checksView.SetHandlers(
		func(suite model.CheckSuite) {
			chkSuite = suite

			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logs, err = c.fetchLogs(chkSuite.Selected)
				if err != nil {
					c.logger.Println(err)
					return
				}
			}

			c.logsView.Load(c.app, view.ModeParseLogs, chkSuite, logs, detailText)
		},
		func(key tcell.Key) {
			c.checksView.Load(c.app, view.ModeChooseCommits, commits, checkRuns, commitSHA)
		},
		func(sha string) {
			commitSHA = sha

			var err error
			checkRuns, err = c.svc.CheckRuns(commitSHA)
			if err != nil {
				c.logger.Println(err)
				return
			}

			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns, commitSHA)
		},
	)

	// logsView
	c.logsView.SetHandlers(
		func(suite model.CheckSuite) {
			chkSuite = suite

			if utils.ShouldShowLogs(chkSuite.Selected) {
				var err error
				logs, err = c.fetchLogs(chkSuite.Selected)
				if err != nil {
					c.logger.Println(err)
					return
				}
			}

			c.logsView.Load(c.app, view.ModeParseLogs, chkSuite, logs, detailText)
		},
		func() {
			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns, commitSHA)
		},
		func(key tcell.Key) {
			logMode = view.ModeParseLogs
			c.logsView.Load(c.app, view.ModeChooseChecks, chkSuite, logs, detailText)
		},
		func() {
			switch logMode {
			case view.ModeParseLogs:
				logMode = view.ModeParseLogsFuller
			default:
				logMode = view.ModeParseLogs
			}

			c.logsView.Load(c.app, logMode, chkSuite, logs, detailText, selection)
		},
		func(id int) {
			logs.Toggle(id)
			selection = view.Selection{Type: view.SelectionTypeID, Value: id}
			c.logsView.Load(c.app, logMode, chkSuite, logs, detailText, selection)
		},
		func(txt string, row int) {
			detailText = txt
			selection = view.Selection{Type: view.SelectionTypeRow, Value: row}
			c.logsView.UpdateDetail(detailText)
		})

	// HANDLE AUTOMATIC EVENTS
	for {
		select {
		// when workflows are done loading
		case <-c.svc.FetchChan:
			c.checksView.Load(c.app, view.ModeChooseChecks, commits, checkRuns)
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
