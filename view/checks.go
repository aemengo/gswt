package view

import (
	"github.com/aemengo/gswt/model"
	"github.com/aemengo/gswt/service"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"sort"
	"strconv"
)

type Checks struct {
	svc *service.Service

	CheckSuiteChan chan model.CheckSuite

	EscapeCheckListChan chan bool
	SelectedCommitChan  chan string
}

func NewChecks(svc *service.Service) *Checks {
	return &Checks{
		svc: svc,

		CheckSuiteChan: make(chan model.CheckSuite),

		EscapeCheckListChan: make(chan bool),
		SelectedCommitChan:  make(chan string),
	}
}

func (c *Checks) Load(app *tview.Application, mode int, commits []*github.RepositoryCommit, checkRunsList *github.ListCheckRunsResults) {
	commitList := c.buildCommitList(commits)
	checkRunsTable := c.buildCheckRunsTable(checkRunsList)

	flex := tview.NewFlex()

	if len(checkRunsList.CheckRuns) == 0 || mode == ModeChooseCommits {
		flex.AddItem(commitList, 0, 1, true)
		flex.AddItem(checkRunsTable, 0, 2, false)
	} else {
		flex.AddItem(commitList, 0, 1, false)
		flex.AddItem(checkRunsTable, 0, 2, true)
	}

	app.SetRoot(flex, true)
}

func (c *Checks) buildCheckRunsTable(checkRunsList *github.ListCheckRunsResults) *tview.Table {
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

	if len(checkRuns) == 0 {
		table.SetCell(0, 0,
			tview.NewTableCell("").
				SetSelectable(false))

		table.SetCell(0, 1,
			tview.NewTableCell("No data found for this commit.").
				SetTextColor(tcell.ColorDarkGray).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false))
	}

	sort.Slice(checkRuns, func(i, j int) bool {
		return *checkRuns[i].CheckSuite.ID < *checkRuns[j].CheckSuite.ID
	})

	var (
		mem             []int64
		row             = 0
		checkRowMapping = map[int]*github.CheckRun{}
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
		runTextColor := tcell.ColorDarkGray
		runSelectable := false

		if c.svc.HasDataFor(checkRun) {
			runTextColor = tcell.ColorMediumTurquoise
			runSelectable = true
		}

		table.SetCell(row, 0,
			tview.NewTableCell(status).
				SetTextColor(color).
				SetSelectable(false))

		table.SetCell(row, 1,
			tview.NewTableCell(*checkRun.Name).
				SetTextColor(runTextColor).
				SetAttributes(tcell.AttrBold).
				SetSelectable(runSelectable))

		checkRowMapping[row] = checkRun
		row = row + 1
	}

	table.
		SetSelectable(true, false).
		SetDoneFunc(func(key tcell.Key) {
			c.EscapeCheckListChan <- true
		}).
		SetSelectedFunc(func(row, column int) {
			selected := checkRowMapping[row]

			c.CheckSuiteChan <- model.CheckSuite{
				All:      matchesCheckSuite(checkRunsList, selected),
				Selected: selected,
			}
		})

	return table
}

func (c *Checks) buildCommitList(commits []*github.RepositoryCommit) *tview.List {
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
			c.listItemSelectedFunc(*commit.SHA),
		)
	}

	return list
}

func (c *Checks) listItemSelectedFunc(sha string) func() {
	return func() {
		c.SelectedCommitChan <- sha
	}
}

func checkStatus(check *github.CheckRun) (string, tcell.Color) {
	switch check.GetStatus() {
	case "completed":
		switch *check.Conclusion {
		case "success":
			return "✔︎", tcell.ColorForestGreen
		case "skipped":
			return "•", tcell.ColorGray
		default:
			return "✘", tcell.ColorIndianRed
		}
	default:
		return "•", tcell.ColorYellow
	}
}

func matchesCheckSuite(checkRunsList *github.ListCheckRunsResults, selected *github.CheckRun) []*github.CheckRun {
	var result []*github.CheckRun
	for _, item := range checkRunsList.CheckRuns {
		if *item.CheckSuite.ID == *selected.CheckSuite.ID {
			result = append(result, item)
		}
	}
	return result
}
