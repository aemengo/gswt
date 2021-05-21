package view

import (
	"fmt"
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
		if step.Success {
			table.SetCell(row, 0,
				tview.NewTableCell("").
					SetSelectable(false))
		} else {
			table.SetCell(row, 0,
				tview.NewTableCell("✘").
					SetTextColor(tcell.ColorIndianRed).
					SetSelectable(false))
		}

		var title = " ► " + tview.TranslateANSI(cyan.Sprint(step.Title))
		if step.Selected {
			title = " ▼ " + tview.TranslateANSI(cyan.Sprint(step.Title))
		}

		table.SetCell(row, 1,
			tview.NewTableCell(title).
				SetTextColor(tcell.ColorDarkGray).
				SetAttributes(tcell.AttrBold).
				SetSelectable(true))

		rowIDMapping[row] = step.ID
		idRowMapping[step.ID] = row
		row = row + 1

		if step.Selected {
			for i, line := range step.Lines {
				txt := fmt.Sprintf("   %s %s",
					tview.TranslateANSI(boldYellow.Sprint(strconv.Itoa(i+1))),
					line)

				//TODO: reconsider how this is implemented
				if !step.Success && i == len(step.Lines) - 1 {
					txt = fmt.Sprintf("   %s %s",
						tview.TranslateANSI(boldYellow.Sprint(strconv.Itoa(i+1))),
						tview.TranslateANSI(boldRed.Sprint(line)))
				}

				table.SetCell(row, 0,
					tview.NewTableCell("").
						SetSelectable(false))

				table.SetCell(row, 1,
					tview.NewTableCell(txt).
						SetTextColor(tcell.ColorDarkGray).
						SetSelectable(true))

				row = row + 1
			}
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

func rowSelectedFunc(rowIDMapping map[int]int, selectedHandler func(id int)) func(row, column int) {
	return func(row, column int) {
		id, ok := rowIDMapping[row]
		if ok {
			selectedHandler(id)
		}
	}
}
