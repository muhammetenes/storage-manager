package config

import (
	"main/handlers"
)

var Port = 1324

type AwsConfig struct {
	AwsId        string
	AwsSecretKey string
	AwsRegion    string
}

type Config struct {
	Status    bool
	Provider  handlers.Handler
	AwsConfig AwsConfig
}

var Conf Config

func (c *Config) Init(s bool, h handlers.Handler, id string, secretKey string, region string) {
	c.Status = s
	c.Provider = h
	c.AwsConfig = AwsConfig{
		AwsId:        id,
		AwsSecretKey: secretKey,
		AwsRegion:    region,
	}
}

func (c *Config) DestroyConfig() {
	c.Status = false
	c.AwsConfig.AwsId = ""
	c.AwsConfig.AwsSecretKey = ""
	c.AwsConfig.AwsRegion = ""
}
