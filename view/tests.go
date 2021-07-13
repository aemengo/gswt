package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

type Tests struct {
	enterHandler            func()
	selectedHandler         func(id int)
	selectionChangedHandler func(txt string, row int)
	statusBar               *tview.TextView
	detailTV                *tview.TextView
}

func NewTests() *Tests {
	return &Tests{
		enterHandler:            func() {},
		selectedHandler:         func(id int) {},
		selectionChangedHandler: func(txt string, row int) {},
	}
}

func (v *Tests) Load(app *tview.Application, logs model.Logs, mode int, displayMode int, testDuration time.Duration, detailText string, selectedRows ...Selection) {
	statusBar := v.buildStatusBar()
	table := v.buildTestsTable(logs, displayMode, selectedRows...)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(table, 0, 1, true)

	switch displayMode {
	case ModeParseTestsFuller:
		detailTV := v.buildDetailTextView()

		v.detailTV = detailTV
		v.UpdateDetail(detailText)
		flex.AddItem(detailTV, 5, 0, false)
	default:
		v.detailTV = nil
	}

	v.statusBar = statusBar
	v.UpdateStatus(mode, logs, testDuration)
	app.SetRoot(flex, true)
}

func (v *Tests) SetHandlers(enterHandler func(), selectedHandler func(id int), selectionChangedHandler func(txt string, row int)) {
	v.enterHandler = enterHandler
	v.selectedHandler = selectedHandler
	v.selectionChangedHandler = selectionChangedHandler
}

func (v *Tests) UpdateStatus(mode int, logs model.Logs, duration time.Duration) {
	if v.statusBar == nil {
		return
	}

	if mode == ModeParseTestsRunning {
		v.statusBar.SetText(fmt.Sprintf("Running %s... (%s)", testsCount(logs), duration))
	} else {
		v.statusBar.SetText(fmt.Sprintf("Completed %s! (%s)", testsCount(logs), duration))
	}
}

func (v *Tests) UpdateDetail(txt string) {
	if v.detailTV == nil {
		return
	}

	v.detailTV.SetText(txt)
}

func (v *Tests) buildStatusBar() *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)
	return tv
}

func (v *Tests) buildDetailTextView() *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)
	return tv
}

func (v *Tests) buildTestsTable(logs model.Logs, mode int, selectedRows ...Selection) *tview.Table {
	return logsDetailView(
		logs,
		func(key tcell.Key) {},
		v.selectedHandler,
		v.enterHandler,
		v.selectionChangedHandler,
		selectedRows...)
}

func testsCount(logs model.Logs) string {
	count := logs.TestCount()
	if count == 1 {
		return "1 test"
	}

	return fmt.Sprintf("%d tests", count)
}
