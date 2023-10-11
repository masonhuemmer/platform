package config

type ConfigurationDirector struct {
	builder IConfigurationBuilder
}

func NewDirector(b IConfigurationBuilder) *ConfigurationDirector {
	return &ConfigurationDirector{
		builder: b,
	}
}

func (d *ConfigurationDirector) Build(endpoint string, labels []string) *Configuration {
	d.builder.setClient(endpoint)
	config := d.builder.getConfig(labels)
	return config

}
