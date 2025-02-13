package main

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type CDNConfig struct {
	AWS_key      string
	AWS_secret   string
	AWS_endpoint string
	AWS_region   string
	AWS_bucket   string
}

var config CDNConfig

func loadConfig() bool {
	if _, err := os.Stat("config.toml"); err == nil {
		content, fileerr := os.ReadFile("config.toml")
		if fileerr != nil {
			panic(fileerr)
		}
		configerr := toml.Unmarshal(content, &config)
		if configerr != nil {
			panic(configerr)
		}
		return true
	} else {
		// Create config and return
		config = CDNConfig{}
		b, configerr := toml.Marshal(config)
		if configerr != nil {
			panic(configerr)
		}
		fmt.Println(string(b))
		fileerr := os.WriteFile("config.toml", b, 0600)
		if fileerr != nil {
			panic(fileerr)
		}
		return false
	}
}
