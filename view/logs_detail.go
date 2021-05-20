package view

import (
	"github.com/aemengo/gswt/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func logsDetailView(logs model.Logs, handler func(key tcell.Key)) *tview.Table {
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

	row := 0
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

		table.SetCell(row, 1,
			tview.NewTableCell(" ► "+tview.TranslateANSI(cyan.Sprint(step.Title))).
				SetTextColor(tcell.ColorDarkGray).
				SetAttributes(tcell.AttrBold).
				SetSelectable(true))

		row = row + 1
	}

	table.
		SetSelectable(true, false).
		SetDoneFunc(handler).
		SetSelectedFunc(func(row, column int) {
			//selected := checkRowMapping[row]
			//
			//c.CheckSuiteChan <- model.CheckSuite{
			//	All:      matchesCheckSuite(checkRunsList, selected),
			//	Selected: selected,
			//}
		})

	return table
}
