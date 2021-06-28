package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

type Tests struct {
	SelectedStepChan chan int
	startTime        time.Time
}

func NewTests() *Tests {
	return &Tests{
		SelectedStepChan: make(chan int),
		startTime:        time.Now(),
	}
}

func (v *Tests) Load(app *tview.Application, logs model.Logs, mode int) {
	statusBar := v.buildStatusBar(mode, func() { app.Draw() })
	table := v.buildTestsTable(logs)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(table, 0, 1, mode == ModeParseTestsFinished)

	app.SetRoot(flex, true)
}

func (v *Tests) buildStatusBar(mode int, handler func()) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).
		SetChangedFunc(handler).
		SetTextAlign(tview.AlignRight).
		SetTextColor(tcell.ColorLightGray).
		SetBackgroundColor(viewBackgroundColor)

	duration := time.Now().Sub(v.startTime)

	if mode == ModeParseTestsRunning {
		tv.SetText(fmt.Sprintf("Running... (%s)", duration))
	} else {
		tv.SetText(fmt.Sprintf("Completed! (%s)", duration))
	}

	return tv
}

func (v *Tests) buildTestsTable(logs model.Logs) *tview.Table {
	escHandler := func(key tcell.Key) {}
	enterHandler := func() {}
	selectedHandler := func(id int) {
		v.SelectedStepChan <- id
	}

	return logsDetailView(
		logs,
		escHandler,
		selectedHandler,
		enterHandler)
}
