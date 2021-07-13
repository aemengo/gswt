package view

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
)

type Logs struct {
	checkSuiteHandler       func(suite model.CheckSuite)
	escLogsHandler          func()
	escLogsDetailHandler    func(key tcell.Key)
	enterHandler            func()
	selectedHandler         func(id int)
	selectionChangedHandler func(txt string, row int)

	detailTV *tview.TextView
}

func NewLogs() *Logs {
	return &Logs{
		checkSuiteHandler:       func(suite model.CheckSuite) {},
		escLogsHandler:          func() {},
		escLogsDetailHandler:    func(key tcell.Key) {},
		enterHandler:            func() {},
		selectedHandler:         func(id int) {},
		selectionChangedHandler: func(txt string, row int) {},
	}
}

func (c *Logs) Load(app *tview.Application, mode int, checks model.CheckSuite, logs model.Logs, detailText string, selectedRows ...Selection) {
	commitList := c.buildTasksList(checks)
	logsDetail := c.buildLogs(mode, checks, logs, selectedRows...)

	flex := tview.NewFlex()

	switch mode {
	case ModeChooseChecks:
		flex.AddItem(commitList, 0, 1, true)
		flex.AddItem(logsDetail, 0, 4, false)
	default:
		flex.AddItem(commitList, 0, 1, false)
		flex.AddItem(logsDetail, 0, 4, true)
	}

	switch mode {
	case ModeParseLogsFuller:
		detailTV := c.buildDetailTextView()

		c.detailTV = detailTV
		c.UpdateDetail(detailText)
		wrapperFlex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(flex, 0, 1, true).
			AddItem(detailTV, 5, 0, false)

		app.SetRoot(wrapperFlex, true)
	default:
		c.detailTV = nil
		app.SetRoot(flex, true)
	}
}

func (c *Logs) SetHandlers(checkSuiteHandler func(suite model.CheckSuite), escLogsHandler func(), escLogsDetailHandler func(key tcell.Key), enterHandler func(), selectedHandler func(id int), selectionChangedHandler func(txt string, row int)) {
	c.checkSuiteHandler = checkSuiteHandler
	c.escLogsHandler = escLogsHandler
	c.escLogsDetailHandler = escLogsDetailHandler
	c.enterHandler = enterHandler
	c.selectedHandler = selectedHandler
	c.selectionChangedHandler = selectionChangedHandler
}

func (c *Logs) UpdateDetail(txt string) {
	if c.detailTV == nil {
		return
	}

	c.detailTV.SetText(txt)
}

func (c *Logs) buildLogs(mode int, checks model.CheckSuite, logs model.Logs, selectedRows ...Selection) tview.Primitive {
	if utils.ShouldShowLogs(checks.Selected) {
		return logsDetailView(
			logs,
			c.escLogsDetailHandler,
			c.selectedHandler,
			c.enterHandler,
			c.selectionChangedHandler,
			selectedRows...)
	} else {
		txtView := tview.NewTextView()
		txtView.
			SetDynamicColors(true).
			SetText("[::b]Sorry, logs only supported for 'success' or 'failure' runs").
			SetTextColor(tcell.ColorDarkGray).
			SetDoneFunc(c.escLogsDetailHandler).
			SetBorder(true).
			SetTitleColor(tcell.ColorDimGray).
			SetBorderPadding(1, 1, 2, 2).
			SetBorderColor(tcell.ColorDimGray).
			SetBorderAttributes(tcell.AttrBold).
			SetBackgroundColor(viewBackgroundColor)
		return txtView
	}
}

func (c *Logs) buildTasksList(checks model.CheckSuite) *tview.List {
	list := tview.NewList()
	list.
		SetMainTextColor(tcell.ColorMediumTurquoise).
		SetSelectedTextColor(tcell.ColorMediumTurquoise).
		SetSelectedBackgroundColor(tcell.ColorDarkSlateGray).
		SetSecondaryTextColor(tcell.ColorDimGray).
		SetTitle("[::b]| tasks |").
		SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetTitleAlign(tview.AlignLeft).
		SetTitleColor(tcell.ColorDimGray).
		SetBorderPadding(1, 1, 2, 2).
		SetBorderColor(tcell.ColorBlack).
		SetBackgroundColor(viewBackgroundColor)

	selectedIndex := 0
	for i, chk := range checks.All {
		if chk.ID == checks.Selected.ID {
			selectedIndex = i
		}

		list.AddItem(
			*chk.Name,
			checkRunStatus(chk),
			0,
			c.listItemSelectedFunc(checks, chk),
		)
	}

	list.
		SetCurrentItem(selectedIndex).
		SetDoneFunc(c.escLogsHandler)
	return list
}

func (c *Logs) buildDetailTextView() *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)
	return tv
}

func (c *Logs) listItemSelectedFunc(chkSuite model.CheckSuite, selected *github.CheckRun) func() {
	return func() {
		suite := model.CheckSuite{All: chkSuite.All, Selected: selected}
		c.checkSuiteHandler(suite)
	}
}

func checkRunStatus(check *github.CheckRun) string {
	switch check.GetStatus() {
	case "completed":
		switch check.GetConclusion() {
		case "success":
			return "[green]✔︎ [-]" + check.GetConclusion()
		case "skipped":
			return "• " + check.GetConclusion()
		default:
			return "[red]✘ [-]" + check.GetConclusion()
		}
	default:
		return "[yellow]• [-]" + check.GetStatus()
	}
}
