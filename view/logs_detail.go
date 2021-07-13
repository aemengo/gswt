package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"regexp"
	"strings"
)

const (
	SelectionTypeID = iota
	SelectionTypeRow
)

type Selection struct {
	Type  int
	Value int
}

func logsDetailView(logs model.Logs, escHandler func(key tcell.Key), selectedHandler func(id int), enterHandler func(), selectionChangedHandler func(txt string, row int), selections ...Selection) *tview.Table {
	var (
		row          = 0
		rowIDMapping = map[int]int{}
		idRowMapping = map[int]int{}
		table        = tview.NewTable()
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

	if len(selections) != 0 {
		s := selections[0]

		switch s.Type {
		case SelectionTypeID:
			r := idRowMapping[s.Value]
			table.Select(r, 1)
		case SelectionTypeRow:
			table.Select(s.Value, 1)
		}
	} else {
		if row != 0 {
			table.Select(row-1, 1)
		}
	}

	style := tcell.StyleDefault.
		Foreground(tcell.ColorMediumTurquoise).
		Background(tcell.ColorDarkSlateGray).
		Attributes(tcell.AttrBold)

	table.
		SetSelectable(true, false).
		SetDoneFunc(escHandler).
		SetSelectedFunc(rowSelectedFunc(rowIDMapping, selectedHandler, enterHandler)).

		// The following handler must come after the table.Select() above
		// otherwise an infinite loop happens
		SetSelectionChangedFunc(selectionChangedFunc(table, selectionChangedHandler)).
		SetSelectedStyle(style).
		SetBorder(true).
		SetTitleColor(tcell.ColorDimGray).
		SetBorderPadding(1, 1, 2, 2).
		SetBorderColor(tcell.ColorDimGray).
		SetBorderAttributes(tcell.AttrBold).
		SetBackgroundColor(viewBackgroundColor)
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

	txt := fmt.Sprintf("[mediumturquoise]Step %d: %s", index+1, truncate(step.Title))

	table.SetCell(*row, 1,
		tview.NewTableCell(icon+txt).
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

	var (
		diffRemoveRegex = regexp.MustCompile(`^\s*-`)
		diffAddRegex    = regexp.MustCompile(`^\s*\+`)
		goFileRegex     = regexp.MustCompile(`(\S+\.go:\d+:)`)
	)

	for _, line := range run.Lines {
		txt := goFileRegex.ReplaceAllString(line, "[mediumturquoise]$1[-]")

		switch {
		case diffRemoveRegex.MatchString(line):
			txt = "[red]" + txt
		case diffAddRegex.MatchString(line):
			txt = "[green]" + txt
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(tview.TranslateANSI("        "+txt)).
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
			tview.NewTableCell(icon+strings.ReplaceAll(tr.Name, "_", " ")).
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

		txt := failureRegex.ReplaceAllString(ts.Title, "[red::b]$1[-:-:-]")

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(icon+fmt.Sprintf("[darkgray]"+txt+"[darkgray]")).
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

	diffRemoveRegex := regexp.MustCompile(`^\s*-`)
	diffAddRegex := regexp.MustCompile(`^\s*\+`)

	for i, line := range step.Lines {
		var (
			txt    string
			prefix = fmt.Sprintf("   [yellow::b]%d[-:-:-] ", i+1)
		)

		switch {
		case diffRemoveRegex.MatchString(line):
			txt = prefix + "[red]" + line
		case diffAddRegex.MatchString(line):
			txt = prefix + "[green]" + line
		case !step.Success && i == len(step.Lines)-1:
			txt = prefix + "[red::b]" + line
		default:
			txt = prefix + line
		}

		table.SetCell(*row, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(*row, 1,
			tview.NewTableCell(tview.TranslateANSI(txt)).
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

func selectionChangedFunc(table *tview.Table, selectionChangedHandler func(txt string, row int)) func(row, column int) {
	return func(row, column int) {
		txt := table.GetCell(row, column).Text
		selectionChangedHandler(txt, row)
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
