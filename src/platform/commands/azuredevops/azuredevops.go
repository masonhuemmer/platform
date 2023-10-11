package azuredevops_command

import (
	"embed"
	"fmt"
	"log"
	config "main/interfaces/configuration"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

//go:embed bash/*
var scripts embed.FS

func get_help_text(category string) string {

	var (
		help string
	)

	switch category {
	case "azuredevops":
		help = `Usage: platform {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} <subcommand> [args]

	The available commands for Azure DevOps are listed below.

Options:
	{{range .Subcommands }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "scale":
		help = `Usage: platform azuredevops {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Configures the number of agents in your Agent Pool that are allowed
	to sit idle waiting for additional jobs to be triggered in Azure DevOps.
	This sub-command is primarily used by the Ephemeral-Infrastructure pipeline
	from the apps project in Azure DevOps.

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
		organization    string
		pat_token       string
		agent_pool_name string
		idle_agents     int
	)

	command := &cli.Command{
		Name:  "azuredevops",
		Usage: "Used for managing Azure DevOps (dev.azure.com)",
		Subcommands: []*cli.Command{
			{
				Name:  "scale",
				Usage: "Scales the number of agents allowed to sit idle in your Agent Pool",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Usage:       "The name of the agent pool in Azure DevOps",
						Destination: &agent_pool_name,
						Required:    true,
					},
					&cli.IntFlag{
						Name:        "agents",
						Usage:       "The desired number of agents running idle in your agent pool",
						Destination: &idle_agents,
						Required:    true,
					},
				},
				Action: func(ctx *cli.Context) error {

					fmt.Println("##[group] Azure DevOps")
					// App Config Store
					endpoint := "https://my-app.azconfig.io"

					// Create Labels from Coommands
					var labels []string
					full_command := strings.Split(ctx.Command.HelpName, " ")
					labels = append(
						labels,
						full_command[0]+"-"+full_command[1], // e.g. platform-preview
						full_command[0]+"-"+full_command[1]+"-"+full_command[2], // e.g. platform-preview-start
					)

					configBuilder := config.GetBuilder("azconfig.io")
					configDirector := config.NewDirector(configBuilder)
					configmap := configDirector.Build(endpoint, labels)

					organization = configmap.List["ADO_ORGANIZATION"].Value
					pat_token = configmap.List["ADO_PAT_TOKEN"].Value

					connect, err := scripts.ReadFile("bash/scale.sh")
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
						organization,
						pat_token,
						agent_pool_name,
						strconv.Itoa(idle_agents),
					).Output()

					if err != nil {
						fmt.Printf("error %s\n", err)
					}

					output := string(cmd)
					fmt.Println(output)
					defer os.Remove(file.Name())

					fmt.Println("##[endgroup]")
					return nil
				},
				CustomHelpTemplate: get_help_text("scale"),
				HideHelpCommand:    true,
			},
		},
		CustomHelpTemplate: get_help_text("azuredevops"),
		HideHelpCommand:    true,
	}

	return command
}
