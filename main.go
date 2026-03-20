package main

import (
	"os"

	"github.com/regaw-leinad/kroger-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
