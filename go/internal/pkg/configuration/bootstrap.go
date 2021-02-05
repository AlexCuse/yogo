package configuration

import "github.com/spf13/viper"

// Unmarshals the config into a struct
func Unmarshal(cfg interface{}) error {
	viper.SetEnvPrefix("yogo")
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetConfigName("configuration")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(cfg)
}
