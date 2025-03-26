package logx

type Config struct {
	// Level is the log level to use. Default is "info".
	Level string `json:"level" yaml:"level" mapstructure:"level"`
}
