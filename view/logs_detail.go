package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"regexp"
	"strconv"
	"strings"
)

func logsDetailView(logs model.Logs, escHandler func(key tcell.Key), selectedHandler func(id int), enterHandler func(), selectedIDs ...int) *tview.Table {
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

	for index, step := range logs {
		showTitleLine(table, step, index, &row, rowIDMapping, idRowMapping)

		if step.Selected {
			if step.IsTest() {
				showTestSuites(table, step, &row, rowIDMapping, idRowMapping)
				continue
			}

			showLogLines(table, step, &row)
		}
	}

	table.
		SetSelectable(true, false).
		SetDoneFunc(escHandler).
		SetSelectedFunc(rowSelectedFunc(rowIDMapping, selectedHandler, enterHandler))

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

func showTitleLine(table *tview.Table, step model.Step, index int, row *int, rowIDMapping map[int]int, idRowMapping map[int]int) {
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

	var icon = " ► "
	if step.Selected {
		icon = " ▼ "
	}

	txt := cyan.Sprintf("Step %d: %s", index, truncate(step.Title))

	table.SetCell(*row, 1,
		tview.NewTableCell(icon+tview.TranslateANSI(txt)).
			SetTextColor(tcell.ColorDarkGray).
			SetAttributes(tcell.AttrBold).
			SetSelectable(true))

	rowIDMapping[*row] = step.ID
	idRowMapping[step.ID] = *row
	*row = *row + 1
}

func showTestLogLines(table *tview.Table, run model.TestRun, row *int) {
	if len(run.Lines) == 0 {
		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell("        ✘︎").
				SetTextColor(tcell.ColorIndianRed).
				SetSelectable(true))

		*row = *row + 1
		return
	}

	goFileRegex := regexp.MustCompile(`(\S+\.go:\d+:)`)
	errRegex := regexp.MustCompile(`(?i)^\s+error:`)

	for i, line := range run.Lines {

		txt := tview.TranslateANSI(goFileRegex.ReplaceAllString(line, cyan.Sprint("$1")))

		if i == len(run.Lines)-1 {
			if errRegex.MatchString(txt) {
				txt = "[red]" + txt
			}
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell("        "+txt).
				SetTextColor(tcell.ColorDarkGray).
				SetSelectable(true))

		*row = *row + 1
	}
}

func showTestRuns(table *tview.Table, suite model.TestSuite, row *int, rowIDMapping map[int]int, idRowMapping map[int]int) {
	failedTestRuns := suite.FailedTestRuns()

	for _, tr := range failedTestRuns {
		var icon = "      ► "
		if tr.Selected {
			icon = "      ▼ "
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(icon+tr.Name).
				SetTextColor(tcell.ColorLightGray).
				SetSelectable(true))

		rowIDMapping[*row] = tr.ID
		idRowMapping[tr.ID] = *row
		*row = *row + 1

		if tr.Selected {
			showTestLogLines(table, tr, row)
		}
	}
}

func showTestSuites(table *tview.Table, step model.Step, row *int, rowIDMapping map[int]int, idRowMapping map[int]int) {
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
		var icon = "   ► "
		if ts.Selected {
			icon = "   ▼ "
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(icon+tview.TranslateANSI(failureRegex.ReplaceAllString(ts.Title, boldRed.Sprint("$1")))).
				SetTextColor(tcell.ColorDarkGray).
				SetSelectable(true))

		rowIDMapping[*row] = ts.ID
		idRowMapping[ts.ID] = *row
		*row = *row + 1

		if ts.Selected {
			showTestRuns(table, ts, row, rowIDMapping, idRowMapping)
		}
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

func rowSelectedFunc(rowIDMapping map[int]int, selectedHandler func(id int), enterHandler func()) func(row, column int) {
	return func(row, column int) {
		id, ok := rowIDMapping[row]
		if ok {
			selectedHandler(id)
			return
		}

		enterHandler()
	}
}

func truncate(str string) string {
	str = strings.Split(str, "\n")[0]

	num := 30
	bnoden := str

	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}
