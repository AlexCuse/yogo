package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type Configuration struct {
	IEXToken   string
	IEXBaseURL string
	BrokerURL  string
	QuoteTopic string
	StatsTopic string
	HitTopic   string
	ScanTopic  string
}

func Load(configFile string) (Configuration, error) {

	var cfg Configuration

	// Read config file
	f, err := os.Open(configFile)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return cfg, err
	}
	err = toml.Unmarshal(buf, &cfg)
	return cfg, err
}
