package iac

import config "main/interfaces/configuration"

type IIacBuilder interface {
	createWorkspace(*config.Configuration)
	findWorkspace(*config.Configuration)
	setVariables()
	runWorkspace(string)
	deleteWorkspace()
	getOutput()
	getWorkspace() Iac
}

func GetBuilder(builderType string) IIacBuilder {
	if builderType == "app.terraform.io" {
		return newTfcIacBuilder()
	}

	return nil
}
