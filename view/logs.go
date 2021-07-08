package view

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
)

type Logs struct {
	LogsCheckSuiteChan chan model.CheckSuite

	SelectedStepChan     chan int
	ToggleModeChan       chan int
	EscapeLogsChan       chan bool
	EscapeLogsDetailChan chan bool
}

func NewLogs() *Logs {
	return &Logs{
		LogsCheckSuiteChan: make(chan model.CheckSuite),

		SelectedStepChan:     make(chan int),
		ToggleModeChan:       make(chan int),
		EscapeLogsChan:       make(chan bool),
		EscapeLogsDetailChan: make(chan bool),
	}
}

func (c *Logs) Load(app *tview.Application, mode int, checks model.CheckSuite, logs model.Logs, selectedIDs ...int) {
	commitList := c.buildTasksList(checks)
	logsDetail := c.buildLogs(mode, checks, logs, selectedIDs...)

	flex := tview.NewFlex()

	switch mode {
	case ModeChooseChecks:
		flex.AddItem(commitList, 0, 1, true)
		flex.AddItem(logsDetail, 0, 4, false)
	case ModeParseLogsFullScreen:
		flex.AddItem(logsDetail, 0, 1, true)
	default:
		flex.AddItem(commitList, 0, 1, false)
		flex.AddItem(logsDetail, 0, 4, true)
	}

	app.SetRoot(flex, true)
}

func (c *Logs) buildLogs(mode int, checks model.CheckSuite, logs model.Logs, selectedIDs ...int) tview.Primitive {
	escHandler := func(key tcell.Key) {
		c.EscapeLogsDetailChan <- true
	}

	selectedHandler := func(id int) {
		c.SelectedStepChan <- id
	}

	enterHandler := func() {
		c.ToggleModeChan <- mode
	}

	if utils.ShouldShowLogs(checks.Selected) {
		return logsDetailView(logs, escHandler, selectedHandler, enterHandler, selectedIDs...)
	} else {
		txtView := tview.NewTextView()
		txtView.
			SetDynamicColors(true).
			SetText(tview.TranslateANSI(bold.Sprint("Sorry, logs only supported for 'success' or 'failure' runs"))).
			SetTextColor(tcell.ColorDarkGray).
			SetDoneFunc(escHandler).
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
		SetTitle(tview.TranslateANSI(bold.Sprint("| tasks |"))).
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
		SetDoneFunc(func() {
			c.EscapeLogsChan <- true
		})
	return list
}

func (c *Logs) listItemSelectedFunc(chkSuite model.CheckSuite, selected *github.CheckRun) func() {
	return func() {
		c.LogsCheckSuiteChan <- model.CheckSuite{
			All:      chkSuite.All,
			Selected: selected,
		}
	}
}

func checkRunStatus(check *github.CheckRun) string {
	switch check.GetStatus() {
	case "completed":
		switch check.GetConclusion() {
		case "success":
			return tview.TranslateANSI(green.Sprint("✔︎ ")) + check.GetConclusion()
		case "skipped":
			return "• " + check.GetConclusion()
		default:
			return tview.TranslateANSI(red.Sprint("✘ ")) + check.GetConclusion()
		}
	default:
		return tview.TranslateANSI(yellow.Sprint("• ")) + check.GetStatus()
	}
}
