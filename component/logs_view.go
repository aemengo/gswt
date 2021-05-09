package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
)

type LogsView struct {
	LogsCheckSuiteChan     chan CheckSuite
	EscapeLogsView         chan bool
	EscapeLogsTextViewChan chan bool
}

func NewLogsView() *LogsView {
	return &LogsView{
		LogsCheckSuiteChan:     make(chan CheckSuite),
		EscapeLogsView:         make(chan bool),
		EscapeLogsTextViewChan: make(chan bool),
	}
}

func (c *LogsView) Load(app *tview.Application, mode int, checks CheckSuite, logsPath string) error {
	commitList := c.buildChecksList(checks)
	logTxtView := c.buildLogs(checks, logsPath)

	flex := tview.NewFlex()

	if !shouldShowLogs(checks.Selected) || mode == ModeChooseChecks {
		flex.AddItem(commitList, 0, 1, true)
		flex.AddItem(logTxtView, 0, 2, false)
	} else {
		flex.AddItem(commitList, 0, 1, false)
		flex.AddItem(logTxtView, 0, 2, true)
	}

	app.SetRoot(flex, true)
	return nil
}

func (c *LogsView) buildLogs(checks CheckSuite, logsPath string) *tview.TextView {
	txtView := tview.NewTextView()
	txtView.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitleColor(tcell.ColorDimGray).
		SetBorderPadding(1, 1, 2, 2).
		SetTitle(tview.TranslateANSI(bold.Sprint("| logs |"))).
		SetBorderColor(tcell.ColorDimGray).
		SetBorderAttributes(tcell.AttrBold).
		SetBackgroundColor(viewBackgroundColor)

	if shouldShowLogs(checks.Selected) {
		txtView.
			SetText(tview.TranslateANSI(bold.Sprintln(logsPath))).
			SetTextColor(tcell.ColorDarkGray)
	} else {
		txtView.
			SetText(tview.TranslateANSI(bold.Sprint("Sorry, logs only supported for 'success' or 'failure' runs"))).
			SetTextColor(tcell.ColorDarkGray)
	}

	txtView.SetDoneFunc(func(key tcell.Key) {
		c.EscapeLogsTextViewChan <- true
	})

	return txtView
}

func (c *LogsView) buildChecksList(checks CheckSuite) *tview.List {
	list := tview.NewList()
	list.
		SetMainTextColor(tcell.ColorMediumTurquoise).
		SetSelectedTextColor(tcell.ColorMediumTurquoise).
		SetSelectedBackgroundColor(tcell.ColorDarkSlateGray).
		SetSecondaryTextColor(tcell.ColorDimGray).
		SetTitle(tview.TranslateANSI(bold.Sprint("| commits |"))).
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
			c.EscapeLogsView <- true
		})
	return list
}

func (c *LogsView) listItemSelectedFunc(chkSuite CheckSuite, selected *github.CheckRun) func() {
	return func() {
		c.LogsCheckSuiteChan <- CheckSuite{
			All:      chkSuite.All,
			Selected: selected,
		}
	}
}

func checkRunStatus(check *github.CheckRun) string {
	switch *check.Status {
	case "completed":
		return *check.Conclusion
	default:
		return *check.Status
	}
}

func shouldShowLogs(check *github.CheckRun) bool {
	switch *check.Status {
	case "completed":
		switch *check.Conclusion {
		case "success", "failure":
			return true
		}
	}

	return false
}
