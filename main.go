package main

import (
	"fmt"
	"os"

	"id3v2-tui/internal/app"
)

func main() {
	filePath := ""
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	a := app.NewApp()
	if err := a.Run(filePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
