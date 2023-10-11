package main

import (
	azuredevops_command "main/commands/azuredevops"
	config_command "main/commands/config"
	preview_command "main/commands/preview"
	tailscale_command "main/commands/tailscale"
)

func load_commands(cli_main *DefaultCli) {

	// Default
	cli_main.app.Commands = append(cli_main.app.Commands,
		preview_command.GetCommand(),
		tailscale_command.GetCommand(),
		azuredevops_command.GetCommand(),
		config_command.GetCommand(),
	)
}
