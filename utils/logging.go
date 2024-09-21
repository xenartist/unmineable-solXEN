package utils

import (
	"fmt"

	"github.com/rivo/tview"
)

type LogMessageFunc func(*tview.TextView, string)

func LogMessage(logView *tview.TextView, message string) {
	fmt.Fprintf(logView, "%s\n", message)
	logView.ScrollToEnd()
}
