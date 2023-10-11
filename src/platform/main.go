package main

import (
	"log"
	"os"
)

var (
	Version  = "Dev"
	Revision = "Local"
)

func main() {

	app := get_default_cli(Version, Revision)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
