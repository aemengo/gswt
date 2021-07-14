package view

import "github.com/gdamore/tcell/v2"

const (
	ModeChooseCommits = iota
	ModeChooseChecks
	ModeParseLogs
	ModeParseLogsFuller
	ModeParseTestsRunning
	ModeParseTestsFinished
	ModeParseTests
	ModeParseTestsFuller
)

var (
	viewBackgroundColor = tcell.NewRGBColor(0, 43, 54)
)
