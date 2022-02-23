package config

import (
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SentinelAddress []string
	Databases string
}

type DB struct {
	Name string
	Port string
}

type Configuration struct {
	SentinelAddress []string
	DB []DB
}

func Init() Configuration {
	conf := Config{}
	err := envconfig.Process("tunnel", &conf)
	if err != nil {
		panic(err)
	}
	d:= []DB{}
	err = json.Unmarshal([]byte(conf.Databases), &d)
	if err != nil {
		panic(err)
	}
	cfg := Configuration{
		SentinelAddress: conf.SentinelAddress,
		DB:              d,
	}

	return cfg
}

