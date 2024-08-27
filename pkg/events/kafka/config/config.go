package config

type ConsumerConfig struct {
	Enabled          bool           `json:"enabled" koanf:"enabled" default:"false"`
	GroupID          string         `json:"groupId" koanf:"groupId"`
	Topics           []string       `json:"topics" koanf:"topics"`
	OffsetFromNewest bool           `json:"offsetFromNewest" koanf:"offsetFromNewest" default:"false"`
	ConsumerOutput   ConsumerOutput `json:"output" koanf:"output"`
}

type ConsumerOutput struct {
	Stdout       bool   `json:"stdout" koanf:"stdout" default:"false"`
	FileLocation string `json:"fileLocation" koanf:"fileLocation" default:"consumer.log"`
}

type Config struct {
	Address string `json:"address" koanf:"address" default:"localhost:10000"`
	Debug   bool   `json:"debug" koanf:"debug" default:"false"`
	Kafka   struct {
		Addresses []string `json:"addresses" koanf:"addresses"`
	} `json:"kafkaAddresses" koanf:"kafkaAddresses"`
	Consumer struct {
		Enabled bool     `json:"enabled" koanf:"enabled" default:"false"`
		GroupID string   `json:"groupId" koanf:"groupId" default:"test-group"`
		Topics  []string `json:"topics" koanf:"topics" default:"test-topic"`
		Output  struct {
			Stdout bool   `json:"stdout" koanf:"stdout" default:"false"`
			File   string `json:"file" koanf:"file" default:"consumer.log"`
		} `json:"output"`
	}
}
