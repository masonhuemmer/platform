
# Platform CLI 
CLI to assist Engineers as they take their changes to production. 

## Requirements
* An AzureAD Account
* Must run `az login` before using CLI

## Commands
```
Usage: platform [global options] <subcommand> [command options] [args]
  
The available commands for execution are listed below.

Main commands:

  preview        Used for managing Ephemeral infrastructure
  tailscale      Used for managing Tailscale (tailscale.com)
  azuredevops    Used for managing Azure DevOps (dev.azure.com)
  config         Used for managing Configuration centrally hosted in the Cloud

Global options (use these before the subcommand, if any):

  --help     Show this help output, or the help for a specified subcommand. (default: false) 
  --version  Show the current Platform version (default: false)
```

## Local Development

```
az login
go run *.go
```

## Build

```
GOOS=$(goos) GOARCH=$(goarch) go build -ldflags="-X 'main.Version=$(platform_version)' -X 'main.Revision=$(platform_revision)'" -o $(Build.BinariesDirectory)/platform
```
