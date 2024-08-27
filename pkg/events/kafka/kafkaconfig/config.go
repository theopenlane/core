package kafkaconfig

// Config is the configuration for the Kafka event source
type Config struct {
	// Enabled is a flag to determine if the Kafka event source is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// AppName is the name of the application that is publishing events
	AppName string `json:"appName" koanf:"appName" default:"openlane"`
	// Address is the address of the Kafka broker
	Address string `json:"address" koanf:"address" default:"localhost:10000"`
	// Addresses is a list of addresses of the Kafka brokers
	Addresses []string `json:"addresses" koanf:"addresses" default:"[localhost:10000]"`
	// Debug is a flag to determine if the Kafka client should run in debug mode
	Debug bool `json:"debug" koanf:"debug" default:"false"`
}
