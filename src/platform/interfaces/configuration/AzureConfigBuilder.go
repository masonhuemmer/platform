package config

import (
	"context"
	"log"
	"net/url"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/tidwall/gjson"
)

type AzureConfigBuilder struct {
	Endpoint      string
	Credential    *azidentity.DefaultAzureCredential
	Client        *azappconfig.Client
	Configuration Configuration
}

// AZ Builder Functions
func newAzureConfigBuilder() *AzureConfigBuilder {
	return &AzureConfigBuilder{}
}

func setAzureCredential(b *AzureConfigBuilder) {

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal("Failed to initialize client: ", err)
	}
	b.Credential = credential

}

func getAppConfigClient(b *AzureConfigBuilder) {

	client, err := azappconfig.NewClient(b.Endpoint, b.Credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	b.Client = client

}

func getSecretByUri(credential *azidentity.DefaultAzureCredential, reference string) string {

	u, err := url.Parse(reference)
	if err != nil {
		log.Fatal(err)
	}

	kvUri := "https://" + u.Host
	kvSecret := path.Base(reference)

	client, err := azsecrets.NewClient(kvUri, credential, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get a secret. An empty string version gets the latest version of the secret.
	version := ""
	resp, err := client.GetSecret(context.TODO(), kvSecret, version, nil)
	if err != nil {
		log.Fatalf("failed to get the secret: %v", err)
	}

	// Set Value and Return
	value := *resp.Value
	return value

}

// Core Builder Functions
func (b *AzureConfigBuilder) setClient(endpoint string) {

	// Set App Config Store Endpoint
	b.Endpoint = endpoint

	// Get Cloud Credential
	setAzureCredential(b)

	// Create Client of the App Config Store
	getAppConfigClient(b)

}

func (b *AzureConfigBuilder) getConfig(labels []string) *Configuration {

	var (
		key KeyValue
	)

	configmap := make(map[string]KeyValue)

	for _, label := range labels {

		// Pull in Key Pairs from App Config Store
		revPgr := b.Client.NewListSettingsPager(
			azappconfig.SettingSelector{
				KeyFilter:   to.Ptr("*"),
				LabelFilter: to.Ptr(label),
				Fields:      azappconfig.AllSettingFields(),
			},
			nil)

		if revPgr.More() {
			if revResp, revErr := revPgr.NextPage(context.TODO()); revErr == nil {
				for _, setting := range revResp.Settings {
					key.Name = *setting.Key
					if gjson.Valid(*setting.Value) {
						result := gjson.Get(*setting.Value, "uri")
						if result.Exists() {
							key.Value = getSecretByUri(b.Credential, result.String())
						} else {
							key.Value = *setting.Value
						}
					} else {
						key.Value = *setting.Value
					}
					key.ContentType = *setting.ContentType
					configmap[*setting.Key] = key
				}
			}
		}
	}

	b.Configuration.List = configmap
	return to.Ptr(b.Configuration)
}
