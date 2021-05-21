package model

import "github.com/google/go-github/v35/github"

type CheckSuite struct {
	All      []*github.CheckRun
	Selected *github.CheckRun
}
