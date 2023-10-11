package iac

import (
	"fmt"
	config "main/interfaces/configuration"
)

type IacDirector struct {
	builder IIacBuilder
}

func NewDirector(b IIacBuilder) *IacDirector {
	return &IacDirector{
		builder: b,
	}
}

func (d *IacDirector) Build(config *config.Configuration) Iac {

	fmt.Println("[group]Create Workspace")
	d.builder.createWorkspace(config)
	d.builder.setVariables()
	fmt.Println("[endgroup]")

	return d.builder.getWorkspace()
}

func (d *IacDirector) Run() {

	d.builder.runWorkspace("apply")
	d.builder.getOutput()

}

func (d *IacDirector) Dismantle(config *config.Configuration) {

	d.builder.findWorkspace(config)
	d.builder.runWorkspace("destroy")
	d.builder.deleteWorkspace()

}
