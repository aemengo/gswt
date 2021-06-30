package model

type TestRun struct {
	ID       int
	Name     string
	Success  bool
	Selected bool

	Lines []string
}