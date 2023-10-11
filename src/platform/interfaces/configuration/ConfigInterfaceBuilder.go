package config

type IConfigurationBuilder interface {
	setClient(endpoint string)
	getConfig(labels []string) *Configuration
}

func GetBuilder(builderType string) IConfigurationBuilder {
	if builderType == "azconfig.io" {
		return newAzureConfigBuilder()
	}

	return nil
}
