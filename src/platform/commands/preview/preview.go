package preview_command

import (
	"encoding/json"
	"fmt"
	"log"
	config "main/interfaces/configuration"
	iac "main/interfaces/iac"
	"os"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/urfave/cli/v2"
)

func get_help_text(category string) string {

	var (
		help string
	)

	switch category {
	case "preview":
		help = `Usage: platform {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} <subcommand> [args]

	Builds and tears down infrastructure according to Terraform configuration
	files located in a GitHub Repository.
	
	Platform will require that you 'Sign into Azure' using the Azure CLI to 
	retrieve the required Terraform Cloud API token and any other additional  
	configuration from an Azure App Configuration Store and Azure Key
	Vault.

Options:
	{{range .Subcommands }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "start":
		help = `Usage: platform preview {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Creates infrastructure according to Terraform configuration
	files located in a GitHub Repository.
	
	Platform will require that you 'Sign into Azure' using the Azure CLI to 
	retrieve the required Terraform Cloud API token and any other additional  
	configuration from an Azure App Configuration Store and a Azure Key
	Vault.

Options:
	{{range .VisibleFlags }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}--{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "stop":
		help = `Usage: platform preview {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Tears down infrastructure according to Terraform configuration
	files located in a GitHub Repository.
	
	Platform will require that you 'Sign into Azure' using the Azure CLI to 
	retrieve the required Terraform Cloud API token and any other additional  
	configuration from a private Azure App Configuration Store and Azure Key
	Vault.

Options:
	{{range .VisibleFlags }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}--{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "list":
		help = `Usage: platform preview {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Returns list of previews hosted in Azure using the Resource Graph API.

	Platform will require that you 'Sign into Azure' using the Azure CLI
	before running this command. The command will return active/expired 
	Previews hosted in Azure using the Resource Graph API. For resources
	to be returned, they must be assigned the 'TerraformCloud' and
    'ExpirationDate' tags.

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
		service     string
		environment string
		location    string
		workspace   string
		status      string
	)

	command := &cli.Command{
		Name:  "preview",
		Usage: "Used for managing Ephemeral infrastructure",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Create and approve a generated plan in Terraform Cloud to stand up infrastructure",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "service",
						Usage:       "The name of the service to be provisioned.",
						Destination: &service,
						Required:    true,
						Action: func(ctx *cli.Context, service string) error {

							supported := []string{
								"my-app",
							}

							if !slices.Contains(supported, service) {
								return fmt.Errorf("value '%s' not supported. Allowed Value: %v", service, supported)
							}

							return nil
						},
					},
					&cli.StringFlag{
						Name:        "environment",
						Usage:       "Where the service will be deployed.",
						Destination: &environment,
						Required:    false,
						Value:       "preview",
						Action: func(ctx *cli.Context, environment string) error {
							supported := []string{
								"preview",
							}

							if !slices.Contains(supported, environment) {
								return fmt.Errorf("value '%s' not supported. Allowed Value: %v", environment, supported)
							}

							return nil
						},
					},
					&cli.StringFlag{
						Name:        "location",
						Usage:       "Azure Region",
						Destination: &location,
						Required:    true,
						Action: func(ctx *cli.Context, location string) error {
							supported := []string{
								"southcentralus",
								"centralus",
							}

							if !slices.Contains(supported, location) {
								return fmt.Errorf("value '%s' not supported. Allowed Value: %v", location, supported)
							}
							return nil
						},
					},
				},
				Action: func(ctx *cli.Context) error {

					// App Config Store
					endpoint := os.Getenv("APP_CONFIG_STORE")

					// Create Labels from Coommands
					var labels []string
					full_command := strings.Split(ctx.Command.HelpName, " ")
					labels = append(
						labels,
						full_command[0]+"-"+full_command[1], // e.g. platform-preview
						full_command[0]+"-"+full_command[1]+"-"+full_command[2],             // e.g. platform-preview-start
						full_command[0]+"-"+full_command[1]+"-"+full_command[2]+"-"+service, // e.g. platform-preview-start-service
					)

					configBuilder := config.GetBuilder("azconfig.io")
					configDirector := config.NewDirector(configBuilder)
					configmap := configDirector.Build(endpoint, labels)

					// Append Flags to ConfigMap
					configmap.List["SERVICE"] = config.KeyValue{
						Name:        "service",
						Value:       service,
						ContentType: "text/plain",
					}
					configmap.List["ENVIRONMENT"] = config.KeyValue{
						Name:        "environment",
						Value:       environment,
						ContentType: "text/plain",
					}
					configmap.List["LOCATION"] = config.KeyValue{
						Name:        "location",
						Value:       location,
						ContentType: "text/plain",
					}

					wsBuilder := iac.GetBuilder("app.terraform.io")
					wsDirector := iac.NewDirector(wsBuilder)
					wsDirector.Build(configmap)
					wsDirector.Run()

					return nil
				},
				CustomHelpTemplate: get_help_text("start"),
				HideHelpCommand:    true,
			},
			{
				Name:  "stop",
				Usage: "Create and approve a generated plan in Terraform Cloud to tear down infrastructure",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "workspace",
						Usage:       "Terraform Workspace",
						Destination: &workspace,
						Required:    true,
					},
				},
				Action: func(ctx *cli.Context) error {

					// App Config Store
					endpoint := os.Getenv("APP_CONFIG_STORE")

					// Create Labels from Coommands
					var labels []string
					full_command := strings.Split(ctx.Command.HelpName, " ")
					labels = append(
						labels,
						full_command[0]+"-"+full_command[1], // e.g. platform-preview
						full_command[0]+"-"+full_command[1]+"-"+full_command[2], // e.g. platform-preview-stop
					)

					configBuilder := config.GetBuilder("azconfig.io")
					configDirector := config.NewDirector(configBuilder)
					configmap := configDirector.Build(endpoint, labels)

					// Append Flags to ConfigMap
					configmap.List["TFC_WORKSPACE"] = config.KeyValue{
						Name:        "workspace",
						Value:       workspace,
						ContentType: "text/plain",
					}

					wsBuilder := iac.GetBuilder("app.terraform.io")
					wsDirector := iac.NewDirector(wsBuilder)
					wsDirector.Dismantle(configmap)

					return nil
				},
				CustomHelpTemplate: get_help_text("stop"),
				HideHelpCommand:    true,
			},
			{
				Name:  "list",
				Usage: "Returns list of previews currently hosted in the Cloud.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "status",
						Usage:       "Status of current preview resources in Azure (e.g. active, expired)",
						Destination: &status,
						Required:    false,
						Action: func(ctx *cli.Context, status string) error {
							supported := []string{
								"active",
								"expired",
							}

							if !slices.Contains(supported, status) {
								return fmt.Errorf("value '%s' not supported", status)
							}

							return nil
						},
					},
				},
				Action: func(ctx *cli.Context) error {

					var (
						query string
						data  map[string]interface{}
					)

					credential, err := azidentity.NewDefaultAzureCredential(nil)
					if err != nil {
						log.Fatal("Failed to initialize client: ", err)
					}

					client, err := armresourcegraph.NewClient(credential, nil)
					if err != nil {
						log.Fatal("Failed to initialize credential: ", err)
					}

					switch status {
					case "active":
						query = "Resources | where tags['ExpirationDate']!='' and tags['TerraformCloud']!='' | where unixtime_seconds_todatetime(toint(tags.ExpirationDate)) > now() | project workspace = tags.TerraformCloud"
					case "expired":
						query = "Resources | where tags['ExpirationDate']!='' and tags['TerraformCloud']!='' | where unixtime_seconds_todatetime(toint(tags.ExpirationDate)) < now() | project workspace = tags.TerraformCloud"
					}

					management_group := os.Getenv("AZURE_MANAGEMENT_GROUP")
					res, err := client.Resources(ctx.Context, armresourcegraph.QueryRequest{
						Query: to.Ptr(query),
						ManagementGroups: []*string{
							to.Ptr(management_group)},
					}, nil)
					if err != nil {
						log.Fatalf("failed to finish the request: %v", err)
					}

					if *res.TotalRecords == 0 {
						data = map[string]interface{}{
							"count": 0,
							"data":  []map[string]interface{}{},
						}
					} else {
						data = map[string]interface{}{
							"count": *res.TotalRecords,
							"data":  []map[string]interface{}{},
						}

						if m, ok := res.Data.([]interface{}); ok {
							for _, r := range m {
								items := r.(map[string]interface{})
								data["data"] = append(data["data"].([]map[string]interface{}), items)
							}
						}
					}

					jsonData, err := json.Marshal(data)
					if err != nil {
						log.Fatalf("Error marshaling JSON: %s", err)
					}

					// Print the JSON
					fmt.Println(string(jsonData))
					return nil
				},
				CustomHelpTemplate: get_help_text("list"),
				HideHelpCommand:    true,
			},
		},
		CustomHelpTemplate: get_help_text("preview"),
		HideHelpCommand:    true,
	}

	return command
}
