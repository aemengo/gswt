package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aemengo/gswt/controller"
	"github.com/aemengo/gswt/service"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		//goland:noinspection GoErrorStringFormat
		expectNoError(errors.New("Missing required env var 'GITHUB_TOKEN'"))
	}

	if len(os.Args) != 3 {
		expectNoError(errors.New("[USAGE] gswt <org/repo> <pr-number>"))
	}

	arg1 := strings.Split(os.Args[1], "/")
	if len(arg1) != 2 {
		expectNoError(errors.New("[USAGE] gswt <org/repo> <pr-number>"))
	}

	arg2 := os.Args[2]

	prNum, err := strconv.Atoi(arg2)
	if err != nil {
		expectNoError(fmt.Errorf("failed to parse pr-number: %s: %s", arg2, err))
	}

	var (
		org  = arg1[0]
		repo = arg1[1]

		ctx    = context.Background()
		ts     = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc     = oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)

		svc  = service.New(ctx, client, org, repo, prNum)
		app  = tview.NewApplication()
		ctrl = controller.New(svc, app)
	)

	err = ctrl.Run()
	expectNoError(err)
}

func expectNoError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
