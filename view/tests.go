package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

type Tests struct {
	SelectedStepChan      chan int
	ToggleDisplayModeChan chan int
	UserDidScrollChan     chan TxtMsg
}

func NewTests() *Tests {
	return &Tests{
		SelectedStepChan:      make(chan int),
		ToggleDisplayModeChan: make(chan int),
		UserDidScrollChan:     make(chan TxtMsg),
	}
}

func (v *Tests) Load(app *tview.Application, logs model.Logs, mode int, displayMode int, testDuration time.Duration, detailText string, selectedRows ...Selection) {
	statusBar := v.buildStatusBar(mode, logs, testDuration)
	table := v.buildTestsTable(logs, displayMode, selectedRows...)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(table, 0, 1, true)

	if displayMode == ModeParseTestsFuller {
		detail := v.buildDetailTextView(detailText)

		flex.AddItem(detail, 5, 0, false)
	}

	app.SetRoot(flex, true)
}

func (v *Tests) buildStatusBar(mode int, logs model.Logs, duration time.Duration) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)

	if mode == ModeParseTestsRunning {
		tv.SetText(fmt.Sprintf("Running %s... (%s)", testsCount(logs), duration))
	} else {
		tv.SetText(fmt.Sprintf("Completed %s! (%s)", testsCount(logs), duration))
	}

	return tv
}

func (v *Tests) buildDetailTextView(detailText string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)
	tv.SetText(detailText)
	return tv
}

func (v *Tests) buildTestsTable(logs model.Logs, mode int, selectedRows ...Selection) *tview.Table {
	escHandler := func(key tcell.Key) {}

	enterHandler := func() {
		v.ToggleDisplayModeChan <- mode
	}

	selectedHandler := func(id int) {
		v.SelectedStepChan <- id
	}

	selectionChangedHandler := func(txt string, row int) {
		v.UserDidScrollChan <- TxtMsg{
			Msg: txt,
			Row: row,
		}
	}

	return logsDetailView(
		logs,
		escHandler,
		selectedHandler,
		enterHandler,
		selectionChangedHandler,
		selectedRows...)
}

func testsCount(logs model.Logs) string {
	count := logs.TestCount()
	if count == 1 {
		return "1 test"
	}

	return fmt.Sprintf("%d tests", count)
}
