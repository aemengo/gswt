package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aemengo/gswt/controller"
	"github.com/aemengo/gswt/service"
	"github.com/aemengo/gswt/utils"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/rivo/tview"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	//goland:noinspection GoErrorStringFormat
	expectNoError(errors.New("Missing required env var 'GITHUB_TOKEN'"), token == "")
	expectNoError(errors.New("[USAGE] gswt <org/repo> <pr-number>"), len(os.Args) != 3)

	arg1 := strings.Split(os.Args[1], "/")
	expectNoError(errors.New("[USAGE] gswt <org/repo> <pr-number>"), len(arg1) != 2)

	arg2 := os.Args[2]
	prNum, err := strconv.Atoi(arg2)
	expectNoError(fmt.Errorf("failed to parse pr-number: %s: %s", arg2, err), err != nil)

	dir, err := os.UserHomeDir()
	expectNoError(err)

	dir = filepath.Join(dir, ".gswt")
	err = os.MkdirAll(utils.LogsDir(dir), os.ModePerm)
	expectNoError(err)

	f, err := os.OpenFile(filepath.Join(utils.LogsDir(dir), "gswt.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	expectNoError(err)

	var (
		org  = arg1[0]
		repo = arg1[1]

		ctx    = context.Background()
		ts     = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc     = oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
		logger = log.New(f, "[GH-SWT] ", log.LstdFlags)

		app = tview.NewApplication()
	)

	svc, err := service.New(ctx, client, logger, dir, org, repo, prNum)
	expectNoError(err)

	ctrl := controller.New(svc, app, logger)

	logger.Println("Starting...")
	err = ctrl.Run()
	expectNoError(err)
}

func expectNoError(err error, cond ...bool) {
	if len(cond) != 0 {
		if cond[0] {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		return
	}

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
