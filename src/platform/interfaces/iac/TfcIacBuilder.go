package iac

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	config "main/interfaces/configuration"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-tfe"
)

const letterBytes = "0123456789ABCDEF"

type TfcIacBuilder struct {
	Workspace      Iac
	config         *config.Configuration
	tfc_api_token  string
	org            string
	service        string
	environment    string
	location       string
	build_id       string
	self           *tfe.Workspace
	client         *tfe.Client
	ctx            context.Context
	project        *tfe.Project
	oauth_token_id string
}

// Diagnostic represents a diagnostic type message from Terraform, which is how errors
// are usually represented.
type Diagnostic struct {
	Severity string           `json:"severity"`
	Summary  string           `json:"summary"`
	Detail   string           `json:"detail"`
	Address  string           `json:"address,omitempty"`
	Range    *DiagnosticRange `json:"range,omitempty"`
}

// Pos represents a position in the source code.
type Pos struct {
	// Line is a one-based count for the line in the indicated file.
	Line int `json:"line"`

	// Column is a one-based count of Unicode characters from the start of the line.
	Column int `json:"column"`

	// Byte is a zero-based offset into the indicated file.
	Byte int `json:"byte"`
}

// DiagnosticRange represents the filename and position of the diagnostic subject.
type DiagnosticRange struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

// For full decoding, see https://github.com/hashicorp/terraform/blob/main/internal/command/jsonformat/renderer.go
type JSONLog struct {
	Message    string      `json:"@message"`
	Level      string      `json:"@level"`
	Timestamp  string      `json:"@timestamp"`
	Type       string      `json:"type"`
	Diagnostic *Diagnostic `json:"diagnostic"`
}

// TFC Builder Functions
func BoolPointer(b bool) *bool {
	return &b
}

func newTfcIacBuilder() *TfcIacBuilder {
	return &TfcIacBuilder{}
}

// Helper Functions
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func setClient(b *TfcIacBuilder) {

	config := &tfe.Config{
		Token:             b.tfc_api_token,
		RetryServerErrors: true,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal("Failed to initialize client: ", err)
	}

	ctx := context.Background()

	b.ctx = ctx
	b.client = client
	fmt.Print("##[info] Logged into Terraform Cloud\n")
}

func getProjectPreview(b *TfcIacBuilder) {

	// Check project exists
	pl, pl_err := b.client.Projects.List(b.ctx, b.org, &tfe.ProjectListOptions{
		Name: "preview",
	})

	if pl_err != nil {
		log.Fatal(pl_err)
	}

	if pl.TotalCount == 0 {
		log.Fatal("Project not found.")
	} else {
		b.project = pl.Items[0]
	}
}

func setOAuthToken(b *TfcIacBuilder) {

	b.oauth_token_id = b.config.List["TFC_OAUTH_TOKEN_ID"].Value
	// fmt.Print("##[info] Lookup OAuth Token\n")
	// o, err := b.client.OAuthTokens.List(b.ctx, b.org, &tfe.OAuthTokenListOptions{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// b.oauth_token = o.Items[0]
	// fmt.Print("##[info] OAuth Token '" + b.oauth_token.ID + "' returned\n")
}

func logErrorsOnly(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var jsonLog JSONLog
		err := json.Unmarshal([]byte(scanner.Text()), &jsonLog)
		// It's possible this log is not encoded as JSON at all, so errors will be ignored.
		if err == nil && jsonLog.Level == "error" {
			fmt.Println()
			fmt.Println("--- Error Message")
			fmt.Println(jsonLog.Message)
			fmt.Println("---")
			fmt.Println()
			if jsonLog.Type == "diagnostic" {
				fmt.Println("--- Diagnostic Details")
				fmt.Println(jsonLog.Diagnostic.Detail)
				fmt.Println("---")
				fmt.Println()
			}
		}
	}
}

func logRunErrors(ctx context.Context, client *tfe.Client, run *tfe.Run) {
	var reader io.Reader
	var err error

	if run.Apply != nil && run.Apply.Status == tfe.ApplyErrored {
		log.Printf("Reading apply logs from %q", run.Apply.LogReadURL)
		reader, err = client.Applies.Logs(ctx, run.Apply.ID)
	} else if run.Plan != nil && run.Plan.Status == tfe.PlanErrored {
		log.Printf("Reading apply logs from %q", run.Plan.LogReadURL)
		reader, err = client.Plans.Logs(ctx, run.Plan.ID)
	} else {
		log.Fatal("Failed to find an errored plan or apply.")
	}

	if err != nil {
		log.Fatal("Failed to read error log: ", err)
	}

	logErrorsOnly(reader)
}

func readRun(ctx context.Context, client *tfe.Client, id string) *tfe.Run {
	r, err := client.Runs.ReadWithOptions(ctx, id, &tfe.RunReadOptions{
		Include: []tfe.RunIncludeOpt{tfe.RunApply, tfe.RunPlan},
	})
	if err != nil {
		log.Fatal("Failed to read specified run: ", err)
	}
	return r
}

// Core Builder Functions
func (b *TfcIacBuilder) findWorkspace(config *config.Configuration) {

	// Set Org
	b.org = "my-org"

	// Set Config
	b.config = config

	// Set Additiaonl Variables
	b.tfc_api_token = b.config.List["TFC_API_TOKEN"].Value
	b.Workspace.name = b.config.List["TFC_WORKSPACE"].Value

	// Log into Terraform Cloud
	setClient(b)

	// Find Workspace
	fmt.Print("##[info] Lookup '" + b.Workspace.name + "' workspace\n")
	wl, wl_err := b.client.Workspaces.List(b.ctx, b.org, &tfe.WorkspaceListOptions{
		Search: b.Workspace.name,
	})

	if wl_err != nil {
		log.Fatal(wl_err)
	}

	fmt.Printf("##[info] Found '%d' Workspace(s)\n", wl.TotalCount)
	if wl.TotalCount > 1 {
		log.Fatalf("Error: Expected 1 Workspace. Received '%d' matches. Exiting Script.", wl.TotalCount)
	} else if wl.TotalCount == 1 {
		b.self = wl.Items[0]
		fmt.Print("##[info] Workspace '" + b.self.Name + "' exists\n")
	} else {
		log.Fatalf("##[info] Workspace '" + b.self.Name + "' not found.\n")
	}
}

func (b *TfcIacBuilder) createWorkspace(config *config.Configuration) {

	// Set Org
	b.org = "my-org"

	// Set Config
	b.config = config

	// Required Variables
	b.tfc_api_token = b.config.List["TFC_API_TOKEN"].Value
	b.service = b.config.List["SERVICE"].Value
	b.environment = b.config.List["ENVIRONMENT"].Value
	b.location = b.config.List["LOCATION"].Value
	b.build_id = RandStringBytes(4)
	b.Workspace.name = b.service + "-" + b.build_id + "-" + b.environment + "-" + b.location
	b.Workspace.working_directory = "workspaces/" + b.service + "/" + b.environment + "/" + b.location

	// Log into Terraform Cloud
	setClient(b)

	// Find Workspace
	// fmt.Print("##[info] Lookup '" + b.Workspace.name + "' workspace\n")
	// wl, wl_err := b.client.Workspaces.List(b.ctx, b.org, &tfe.WorkspaceListOptions{
	// 	Search: b.Workspace.name,
	// })

	// if wl_err != nil {
	// 	log.Fatal(wl_err)
	// }

	// fmt.Printf("##[info] Found '%d' Workspace(s)\n", wl.TotalCount)
	// if wl.TotalCount > 1 {
	// 	log.Fatalf("Error: Expected 1 Workspace. Received '%d' matches. Exiting Script.", wl.TotalCount)
	// } else if wl.TotalCount == 1 {
	// 	b.self = wl.Items[0]
	// 	fmt.Print("##[info] Workspace '" + b.self.Name + "' exists\n")
	// } else {

	// Preqreuisites
	getProjectPreview(b)
	setOAuthToken(b)

	// Create a new workspace
	fmt.Printf("##[info] Creating Workspace in '%s' Project\n", b.project.Name)
	wc, err := b.client.Workspaces.Create(b.ctx, b.org, tfe.WorkspaceCreateOptions{
		Name:             tfe.String(b.Workspace.name),
		AllowDestroyPlan: tfe.Bool(true),
		ExecutionMode:    tfe.String("remote"),
		WorkingDirectory: tfe.String(b.Workspace.working_directory),
		Project:          b.project,
		Tags: []*tfe.Tag{
			{Name: b.service},
		},
		VCSRepo: &tfe.VCSRepoOptions{
			Branch:            tfe.String("main"),
			Identifier:        tfe.String("Org/Repo"),
			OAuthTokenID:      tfe.String(b.oauth_token_id),
			IngressSubmodules: tfe.Bool(false),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	b.self = wc
	fmt.Print("##[info] Workspace '" + b.self.Name + "' created\n")
}

func (b *TfcIacBuilder) deleteWorkspace() {

	// Find Workspace
	err := b.client.Workspaces.Delete(b.ctx, b.org, b.Workspace.name)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("##[info] Workspace '" + b.Workspace.name + "' deleted.\n")
}

func (b *TfcIacBuilder) setVariables() {

	type Variable struct {
		Value     string            `json:"value"`
		Sensitive bool              `json:"sensitive"`
		Category  *tfe.CategoryType `json:"category"`
	}

	// Create VarMap
	varmap := make(map[string]interface{})
	varmap["ARM_TENANT_ID"] = Variable{
		Value:     b.config.List["ARM_TENANT_ID"].Value,
		Sensitive: false,
		Category:  tfe.Category("env"),
	}
	varmap["ARM_CLIENT_ID"] = Variable{
		Value:     b.config.List["ARM_CLIENT_ID"].Value,
		Sensitive: false,
		Category:  tfe.Category("env"),
	}
	varmap["ARM_CLIENT_SECRET"] = Variable{
		Value:     b.config.List["ARM_CLIENT_SECRET"].Value,
		Sensitive: true,
		Category:  tfe.Category("env"),
	}
	varmap["BUILD_ID"] = Variable{
		Value:     b.build_id,
		Sensitive: false,
		Category:  tfe.Category("terraform"),
	}

	vl, err := b.client.Variables.List(b.ctx, b.self.ID, &tfe.VariableListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for k := range varmap {
		_, err := b.client.Variables.Create(b.ctx, b.self.ID, tfe.VariableCreateOptions{
			Key:       tfe.String(k),
			Value:     tfe.String(varmap[k].(Variable).Value),
			Category:  varmap[k].(Variable).Category,
			Sensitive: tfe.Bool(varmap[k].(Variable).Sensitive),
		})
		if err != nil {
			for _, arr := range vl.Items {
				if k == arr.Key {
					// Update Values
					_, err := b.client.Variables.Update(b.ctx, b.self.ID, arr.ID, tfe.VariableUpdateOptions{
						Value:     tfe.String(varmap[k].(Variable).Value),
						Category:  varmap[k].(Variable).Category,
						Sensitive: tfe.Bool(varmap[k].(Variable).Sensitive),
					})
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
}

func (b *TfcIacBuilder) runWorkspace(RunType string) {

	var (
		pollInterval = 10 * time.Second
	)

	fmt.Println("[group]Create Workspace Run")

	rc, err := b.client.Runs.Create(b.ctx, tfe.RunCreateOptions{
		Message:   tfe.String("Triggered via SDK"),
		Workspace: b.self,
		IsDestroy: tfe.Bool(strings.ToLower(RunType) == "destroy"),
		AutoApply: tfe.Bool(true),
	})

	if err != nil {
		log.Fatal("##[error] Failed to read specified run: ", err)
	}

	r := readRun(b.ctx, b.client, rc.ID)

poll:
	for {
		<-time.After(pollInterval)

		r := readRun(b.ctx, b.client, r.ID)

		switch r.Status {
		case tfe.RunPlannedAndFinished:
			fmt.Println("##[info] Planned and Finished!")
			break poll
		case tfe.RunApplied:
			fmt.Println("##[info] Run Applied!")
			break poll
		case tfe.RunErrored:
			fmt.Println("##[error] Run had errors!")
			logRunErrors(b.ctx, b.client, r)
			break poll
		default:
			fmt.Printf("##[info] Run status %q...\n", r.Status)
		}
	}

	fmt.Println("[endgroup]")
}

func (b *TfcIacBuilder) getOutput() {

	var (
		default_hostname string
	)

	fmt.Println("[group]Terraform Output")

	sv, err := b.client.StateVersionOutputs.ReadCurrent(b.ctx, b.self.ID)
	if err != nil {
		log.Fatal("##[error] Failed to read specified state version: ", err)
	}

	for _, arr := range sv.Items {
		fmt.Printf("##vso[task.setvariable variable=" + arr.Name + ";isOutput=true]" + arr.Value.(string) + "\n")
		outJson := map[string]interface{}{
			"name":  arr.Name,
			"value": arr.Value,
		}
		if strings.Contains(arr.Name, "default_hostname") {
			default_hostname = arr.Value.(string)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(outJson); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("[endgroup]")

	if len(default_hostname) != 0 {
		fmt.Println("##[section] URL: " + default_hostname)
	}
}

func (b *TfcIacBuilder) getWorkspace() Iac {
	return b.Workspace
}
