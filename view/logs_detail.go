package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"regexp"
	"strconv"
)

func logsDetailView(logs model.Logs, selectedIDs []int, escHandler func(key tcell.Key), selectedHandler func(id int)) *tview.Table {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorMediumTurquoise).
		Background(tcell.ColorDarkSlateGray).
		Attributes(tcell.AttrBold)

	table := tview.NewTable()
	table.
		SetSelectedStyle(style).
		SetBorder(true).
		SetTitleColor(tcell.ColorDimGray).
		SetBorderPadding(1, 1, 2, 2).
		SetBorderColor(tcell.ColorDimGray).
		SetBorderAttributes(tcell.AttrBold).
		SetBackgroundColor(viewBackgroundColor)

	var (
		row          = 0
		rowIDMapping = map[int]int{}
		idRowMapping = map[int]int{}
	)

	for _, step := range logs {
		rowIDMapping[row] = step.ID
		idRowMapping[step.ID] = row

		showTitleLine(table, step, &row)

		if step.Selected {
			if step.IsTest() {
				showTestLines(table, step, &row)
				continue
			}

			showLogLines(table, step, &row)
		}
	}

	table.
		SetSelectable(true, false).
		SetDoneFunc(escHandler).
		SetSelectedFunc(rowSelectedFunc(rowIDMapping, selectedHandler))

	if len(selectedIDs) != 0 {
		id := selectedIDs[0]
		r := idRowMapping[id]
		table.Select(r, 1)
	} else {
		if row != 0 {
			table.Select(row-1, 1)
		}
	}

	return table
}

func showTitleLine(table *tview.Table, step model.Step, row *int) {
	if step.Success {
		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))
	} else {
		table.SetCell(*row, 0,
			tview.NewTableCell("✘").
				SetTextColor(tcell.ColorIndianRed).
				SetSelectable(false))
	}

	var title = " ► " + tview.TranslateANSI(cyan.Sprint(step.Title))
	if step.Selected {
		title = " ▼ " + tview.TranslateANSI(cyan.Sprint(step.Title))
	}

	table.SetCell(*row, 1,
		tview.NewTableCell(title).
			SetTextColor(tcell.ColorDarkGray).
			SetAttributes(tcell.AttrBold).
			SetSelectable(true))

	*row = *row + 1
}

func showTestLines(table *tview.Table, step model.Step, row *int) {
	failedTestSuites := step.FailedTestSuites()
	failureRegex := regexp.MustCompile(`(Failed: \d+)`)

	if len(failedTestSuites) == 0 {
		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell("   ✔︎").
				SetTextColor(tcell.ColorForestGreen).
				SetSelectable(true))

		*row = *row + 1
		return
	}

	for _, ts := range failedTestSuites {
		suiteTitle := "   ► " + tview.TranslateANSI(failureRegex.ReplaceAllString(ts.Title, boldRed.Sprint("$1")))

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(suiteTitle).
				SetTextColor(tcell.ColorDarkGray).
				SetSelectable(true))

		*row = *row + 1
	}
}

func showLogLines(table *tview.Table, step model.Step, row *int) {
	if len(step.Lines) == 0 {
		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell("   ✔︎").
				SetTextColor(tcell.ColorForestGreen).
				SetSelectable(true))

		*row = *row + 1
		return
	}

	for i, line := range step.Lines {
		txt := fmt.Sprintf("   %s %s",
			tview.TranslateANSI(boldYellow.Sprint(strconv.Itoa(i+1))),
			line)

		//TODO: reconsider how this is implemented
		if !step.Success && i == len(step.Lines)-1 {
			txt = fmt.Sprintf("   %s %s",
				tview.TranslateANSI(boldYellow.Sprint(strconv.Itoa(i+1))),
				tview.TranslateANSI(boldRed.Sprint(line)))
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(txt).
				SetTextColor(tcell.ColorDarkGray).
				SetSelectable(true))

		*row = *row + 1
	}
}

func rowSelectedFunc(rowIDMapping map[int]int, selectedHandler func(id int)) func(row, column int) {
	return func(row, column int) {
		id, ok := rowIDMapping[row]
		if ok {
			selectedHandler(id)
		}
	}
}
