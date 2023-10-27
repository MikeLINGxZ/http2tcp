package main

import (
	"github.com/MikeLINGxZ/http2tcp"
	"github.com/MikeLINGxZ/http2tcp/internal/client"
	"github.com/spf13/viper"
)

func main() {
	config, err := initConfig()
	if err != nil {
		panic(err)
	}
	clientHandler := client.NewClient(config)
	err = clientHandler.Run()
	if err != nil {
		panic(err)
	}
}

func initConfig() (*http2tcp.ClientConfig, error) {
	viper.Reset()
	viper.SetConfigName("client")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	config := &http2tcp.ClientConfig{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
