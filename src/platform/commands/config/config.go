package config_command

import (
	"encoding/json"
	"fmt"
	"log"
	config "main/interfaces/configuration"
	"slices"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type KeyValueSlice []config.KeyValue

// Implement the sort.Interface methods for KeyValueSlice
func (k KeyValueSlice) Len() int           { return len(k) }
func (k KeyValueSlice) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }
func (k KeyValueSlice) Less(i, j int) bool { return k[i].Name < k[j].Name }

func get_help_text(category string) string {

	var (
		help string
	)

	switch category {
	case "config":
		help = `Usage: platform {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} <subcommand> [args]

	The available commands for Configuration are listed below.

Options:
	{{range .Subcommands }}
	{{range $index, $option := .Names }}{{if $index}}{{end}}{{$option}}{{end}}{{ "\t\t"}}{{.Usage}}{{end}}{{ "\n" }}
`
	case "list":
		help = `Usage: platform config {{if .VisibleFlags}}[global options]{{end}} {{if .Name}}{{ .Name }}{{end}} [options]

	Returns a list of key-value pairs from an App Configuration store hosted in Azure.
	Once the list is returned, it is then formated based on the output selected. 

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
		location    string
		flag_labels cli.StringSlice
		output      string
	)

	command := &cli.Command{
		Name:  "config",
		Usage: "Used for managing Configuration centrally hosted in the Cloud",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "Retrieve the configuration of the service labels.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "service",
						Usage:       "The name of the Service",
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
						Name:        "location",
						Usage:       "Azure Region",
						Destination: &location,
						Required:    false,
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
					&cli.StringSliceFlag{
						Name:        "label",
						Usage:       "If multiples labels are required, assign multiple label flags \n\t\t\t(e.g. platform config list --service payment-portal --label test --label shared --output env)",
						Destination: &flag_labels,
						Required:    false,
					},
					&cli.StringFlag{
						Name:        "output",
						Usage:       "Output format.  Allowed values: env, json, none, yaml.  Default: json.",
						Destination: &output,
						Value:       "env",
						Required:    false,
						Action: func(ctx *cli.Context, output string) error {

							supported := []string{
								"env",
								"json",
								"yaml",
							}

							if !slices.Contains(supported, output) {
								return fmt.Errorf("value '%s' not supported. Allowed Value: %v", output, supported)
							}

							return nil
						},
					},
				},
				Action: func(ctx *cli.Context) error {

					type JsonOutput struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					}

					type YamlOutput struct {
						Key   string `yaml:"key"`
						Value string `yaml:"value"`
					}

					// App Config Store
					endpoint := "https://my-app.azconfig.io"

					// Add Labels
					var labels []string
					full_command := strings.Split(ctx.Command.HelpName, " ")
					labels = flag_labels.Value()
					labels = append(
						labels,
						full_command[0]+"-"+full_command[1], // e.g. platform-preview
						full_command[0]+"-"+full_command[1]+"-"+full_command[2],             // e.g. platform-preview-start
						full_command[0]+"-"+full_command[1]+"-"+full_command[2]+"-"+service, // e.g. platform-preview-start-service
					)

					configBuilder := config.GetBuilder("azconfig.io")
					configDirector := config.NewDirector(configBuilder)
					configmap := configDirector.Build(endpoint, labels)

					var keyValueSlice KeyValueSlice
					for _, value := range configmap.List {
						keyValueSlice = append(keyValueSlice, value)
					}

					sort.Sort(keyValueSlice)

					switch output {
					case "json":
						json_output := []JsonOutput{}
						for _, key := range keyValueSlice {
							json_output = append(json_output, JsonOutput{
								Key:   key.Name,
								Value: key.Value,
							})
						}
						val, err := json.MarshalIndent(json_output, "", "    ")
						if err != nil {
							log.Fatal("Failed to marshal JSON Output: ", err)
						}
						fmt.Println(string(val))
					case "env":
						for _, key := range keyValueSlice {
							k := strings.Replace(key.Name, ":", "_", 1)
							fmt.Println(k + "=" + key.Value)
						}
					case "yaml":
						yaml_output := []YamlOutput{}
						for _, key := range keyValueSlice {
							yaml_output = append(yaml_output, YamlOutput{
								Key:   key.Name,
								Value: key.Value,
							})
						}
						val, err := yaml.Marshal(yaml_output)
						if err != nil {
							log.Fatal("Failed to marshal JSON Output: ", err)
						}
						fmt.Println(string(val))
					}

					return nil
				},
				CustomHelpTemplate: get_help_text("list"),
				HideHelpCommand:    true,
			},
		},
		CustomHelpTemplate: get_help_text("config"),
		HideHelpCommand:    true,
	}

	return command
}
