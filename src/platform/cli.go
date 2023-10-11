package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
)

type DefaultCli struct {
	name    string
	version string
	app     *cli.App
}

func set_help_text() {

	cli.AppHelpTemplate = `Usage: {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} <subcommand> [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[args]{{end}}
	
The available commands for execution are listed below.
{{if .Commands}}
Main commands:
{{range .Commands}}{{if not .HideHelp}}
{{ "\t"}}{{range $index, $option := .Names }}{{if $index}}{{end}}{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{end}}{{end}}
{{if .VisibleFlags}}
Global options (use these before the subcommand, if any):

{{range $index, $option := .VisibleFlags}}{{if $index}} 
{{end}}{{ "\t" }}{{$option.String}}{{end}}{{end}}{{ "\n" }}
`
}

func get_default_cli(version string, revision string) *cli.App {

	cli_main := &DefaultCli{}
	cli_main.name = "Platform"
	cli_main.version = version
	cli_main.app = &cli.App{
		Name:                 cli_main.name,
		HelpName:             strings.ToLower(cli_main.name),
		Usage:                "a shoe with very thick soles",
		Version:              cli_main.version,
		HideHelpCommand:      true,
		EnableBashCompletion: true,
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "Show the current " + cli_main.name + " version",
	}

	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("%s v%s.%s\non %s_%s\n", cCtx.App.Name, version, revision, runtime.GOOS, runtime.GOARCH)
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:  "help",
		Usage: "Show this help output, or the help for a specified subcommand.",
	}

	set_help_text()
	load_commands(cli_main)

	return cli_main.app
}
