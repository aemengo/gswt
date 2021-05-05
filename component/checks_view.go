package component

import (
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"sort"
)

type ChecksView struct {}

func NewChecksView() *ChecksView {
	return &ChecksView{}
}

func (c *ChecksView) Load(app *tview.Application, commits []*github.RepositoryCommit, checkRunsList *github.ListCheckRunsResults) error {
	flex := tview.NewFlex()
	commitList := tview.NewList()
	commitList.
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
		SetBackgroundColor(tcell.NewRGBColor(0, 43, 54))

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Commit.Committer.Date.After(*commits[j].Commit.Committer.Date)
	})

	for _, commit := range commits {
		commitList.AddItem(
			*commit.SHA,
			humanize.Time(*commit.Commit.Committer.Date),
			0,
			func() {},
		)
	}

	flex.AddItem(commitList, 0,1, true)
	app.SetRoot(flex, true)
	return nil
}

