package main

import (
	"os"

	"pdftool/internal/cmd"
	"pdftool/internal/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
