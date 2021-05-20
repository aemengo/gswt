package view

import (
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
)

var (
	bold   = color.New(color.Bold)
	green  = color.New(color.FgHiGreen)
	red    = color.New(color.FgHiRed)
	yellow = color.New(color.FgHiYellow)
	cyan   = color.New(color.FgHiCyan)

	viewBackgroundColor = tcell.NewRGBColor(0, 43, 54)
)