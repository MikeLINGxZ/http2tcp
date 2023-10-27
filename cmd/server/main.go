package main

import (
	"github.com/MikeLINGxZ/http2tcp"
	"github.com/spf13/viper"
)

func main() {
	config, err := initConfig()
	if err != nil {
		panic(err)
	}
	serverHandler := http2tcp.NewServer(config)
	err = serverHandler.Run()
	if err != nil {
		panic(err)
	}
}

func initConfig() (*http2tcp.ServerConfig, error) {
	viper.Reset()
	viper.SetConfigName("server")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	config := &http2tcp.ServerConfig{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
