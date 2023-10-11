package config

type KeyValue struct {
	Name        string
	Value       string
	ContentType string
}

type Configuration struct {
	List map[string]KeyValue
}
