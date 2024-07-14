package pulse

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Name                 string
		Hostname             string
		Port                 string
		AggregationInterval  time.Duration
		SegmentationInterval time.Duration
		SegmentSizeKB        int
	}
	Database struct {
		Address  string
		Password string
	}
}

func ParseConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.pulse")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	return &cfg, err
}
