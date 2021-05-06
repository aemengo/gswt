package component

import (
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"sort"
	"strconv"
)

type ChecksView struct{}

func NewChecksView() *ChecksView {
	return &ChecksView{}
}

func (c *ChecksView) Load(app *tview.Application, commits []*github.RepositoryCommit, checkRunsList *github.ListCheckRunsResults) error {
	flex := tview.NewFlex()
	commitList := buildCommitList(commits)
	checkRunsTable := buildCheckRunsTable(checkRunsList)

	flex.AddItem(commitList, 0, 1, false)
	flex.AddItem(checkRunsTable, 0, 2, true)

	app.SetRoot(flex, true)
	return nil
}

func buildCheckRunsTable(checkRunsList *github.ListCheckRunsResults) *tview.Table {
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
		SetTitle(tview.TranslateANSI(bold.Sprint("| checks |"))).
		SetBorderColor(tcell.ColorDimGray).
		SetBorderAttributes(tcell.AttrBold).
		SetBackgroundColor(viewBackgroundColor)

	checkRuns := checkRunsList.CheckRuns

	sort.Slice(checkRuns, func(i, j int) bool {
		return *checkRuns[i].CheckSuite.ID < *checkRuns[j].CheckSuite.ID
	})

	var (
		row = 0
		mem []int64
	)

	for _, checkRun := range checkRuns {
		if len(mem) == 0 || mem[len(mem)-1] != *checkRun.CheckSuite.ID {
			mem = append(mem, *checkRun.CheckSuite.ID)

			if row != 0 {
				row = row + 1
			}

			table.SetCell(row, 0,
				tview.NewTableCell("").
					SetSelectable(false))

			table.SetCell(row, 1,
				tview.NewTableCell("Task "+strconv.Itoa(len(mem))).
					SetTextColor(tcell.ColorDarkGray).
					SetAttributes(tcell.AttrBold).
					SetSelectable(false))

			row = row + 1
		}

		status, color := checkStatus(checkRun)

		table.SetCell(row, 0,
			tview.NewTableCell(status).
				SetTextColor(color).
				SetSelectable(false))

		table.SetCell(row, 1,
			tview.NewTableCell(*checkRun.Name).
				SetTextColor(tcell.ColorMediumTurquoise).
				SetAttributes(tcell.AttrBold).
				SetSelectable(true))

		row = row + 1
	}

	if row > 0 {
		table.
			SetSelectable(true, false).
			Select(1, 0)
	}

	return table
}

func buildCommitList(commits []*github.RepositoryCommit) *tview.List {
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

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Commit.Committer.Date.After(*commits[j].Commit.Committer.Date)
	})

	for _, commit := range commits {
		list.AddItem(
			*commit.SHA,
			humanize.Time(*commit.Commit.Committer.Date),
			0,
			func() {},
		)
	}

	return list
}

func checkStatus(check *github.CheckRun) (string, tcell.Color) {
	switch *check.Status {
	case "completed":
		switch *check.Conclusion {
		case "success":
			return "✔", tcell.ColorForestGreen
		case "skipped":
			return "•", tcell.ColorGray
		default:
			return "✘", tcell.ColorIndianRed
		}
	default:
		return "•", tcell.ColorYellow
	}
}
