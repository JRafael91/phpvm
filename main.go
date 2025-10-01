package main

import (
	"os"

	"github.com/yourusername/phpvm/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
