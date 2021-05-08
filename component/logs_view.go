package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
)

type LogsView struct {
	EscapeLogsView chan bool
}

func NewLogsView() *LogsView {
	return &LogsView{
		EscapeLogsView: make(chan bool),
	}
}

func (c *LogsView) Load(app *tview.Application, checks CheckSuite) error {
	commitList := c.buildChecksList(checks)

	flex := tview.NewFlex()
	flex.AddItem(commitList, 0, 1, true)

	app.SetRoot(flex, true)
	return nil
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
			func() {},
		)
	}

	list.
		SetCurrentItem(selectedIndex).
		SetDoneFunc(func() {
			c.EscapeLogsView <- true
		})
	return list
}

func checkRunStatus(check *github.CheckRun) string {
	switch *check.Status {
	case "completed":
		return *check.Conclusion
	default:
		return *check.Status
	}
}
