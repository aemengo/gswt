package main

import (
	"fmt"
	"github.com/aemengo/gswt/controller"
	"github.com/aemengo/gswt/utils"
	"github.com/rivo/tview"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.UserHomeDir()
	expectNoError(err)

	dir = filepath.Join(dir, ".gswt")
	err = os.MkdirAll(utils.LogsDir(dir), os.ModePerm)
	expectNoError(err)

	f, err := os.OpenFile(filepath.Join(utils.LogsDir(dir), "gswt.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	expectNoError(err)

	var (
		logger = log.New(f, "[GO-SWT] ", log.LstdFlags)
		app    = tview.NewApplication()
	)

	ctrl := controller.NewCLController(app, logger, os.Stdin)

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
