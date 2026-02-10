package core

var (
	config *Config
)

type Config struct {
	Eks *EksConfig
}

func GetConfig() *Config {
	return config
}

type EksConfig struct {
	Id string `json:"id" yaml:"id"`
	Az string `json:"az" yaml:"az"`
}
