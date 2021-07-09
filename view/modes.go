package view

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

type TxtMsg struct {
	Msg string
	Row int
}
