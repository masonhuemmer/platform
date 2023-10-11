package tailscale_command

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/urfave/cli/v2"
)

//go:embed bash/*
var scripts embed.FS

func get_help_text(category string) string {

	var (
		help string
	)

	switch category {
	case "tailscale":
		help = `Usage: platform {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} <subcommand> [args]

	The available commands for Tailscale are listed below.

Options:
	{{range .Subcommands }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "install":
		help = `Usage: platform tailscale {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Downloads tailscale and configures agent according to the flags assigned.
	At this moment, Linux is the only supported OS.
	
	Platform will require that you 'Sign into Azure' using the Azure CLI to 
	retrieve the required OAuth Client ID and Secret and any other additional  
	configuration from a private Azure App Configuration Store and Azure Key
	Vault. After configuration has been returned, Platform 
	will download the Tailscale agent and then run Tailscale Up.

Options:
	{{range .VisibleFlags }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}--{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	}

	return help

}

func GetCommand() *cli.Command {

	// Placeholders
	var (
		oauth_client_id     string
		oauth_client_secret string
	)

	command := &cli.Command{
		Name:  "tailscale",
		Usage: "Used for managing Tailscale (tailscale.com)",
		Subcommands: []*cli.Command{
			{
				Name:  "install",
				Usage: "Downloads tailscale and configures the agent",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "client-id",
						Usage:       "Tailscale OAuth Client ID",
						Destination: &oauth_client_id,
						Required:    false,
						Value:       "",
					},
					&cli.StringFlag{
						Name:        "client-secret",
						Usage:       "Tailscale OAuth Client Secret",
						Destination: &oauth_client_secret,
						Required:    false,
						Value:       "",
					},
				},
				Action: func(ctx *cli.Context) error {

					fmt.Println("##[group] Download Tailscale")
					if runtime.GOOS == "linux" {

						connect, err := scripts.ReadFile("bash/connect.sh")
						if err != nil {
							log.Fatal(err)
						}

						file, err := os.CreateTemp("/tmp", "platform.*.sh")
						if err != nil {
							log.Fatal(err)
						}
						file.Write(connect)

						cmd, err := exec.Command(
							"/bin/bash",
							file.Name(),
							oauth_client_id,
							oauth_client_secret,
						).Output()

						if err != nil {
							fmt.Printf("error %s\n", err)
						}

						output := string(cmd)
						fmt.Println(output)
						defer os.Remove(file.Name())

					}
					fmt.Println("##[endgroup]")
					return nil
				},
				CustomHelpTemplate: get_help_text("install"),
				HideHelpCommand:    true,
			},
		},
		CustomHelpTemplate: get_help_text("tailscale"),
		HideHelpCommand:    true,
	}

	return command
}
