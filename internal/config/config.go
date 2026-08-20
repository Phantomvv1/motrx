package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type Server struct {
	Address string `json:"address"`
}

type HealthCheck struct {
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type Retry struct {
	MaxTries int `json:"max_tries"`
}

type Config struct {
	Servers                []Server    `json:"servers"`
	LoadBalancingAlgorithm string      `json:"load_balancing_algorithm"`
	HealthCheck            HealthCheck `json:"health_check"`
	Retry                  Retry       `json:"retry"`
}

func ParseConfig(path string) (*Config, error) {
	fileName := path
	index := strings.LastIndex(path, "/")
	if index != -1 {
		fileName = path[index+1:]
	}

	parts := strings.Split(fileName, ".")
	if len(parts) == 1 {
		return nil, errors.New("Error: there is no file format")
	}

	if parts[1] != "json" && parts[1] != "yaml" && parts[1] != "yml" {
		return nil, errors.New("Error: invalid file format")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf := &Config{}
	if parts[1] == "json" {
		err := json.NewDecoder(file).Decode(conf)
		if err != nil {
			return nil, err
		}
	}

	if parts[1] == "yaml" || parts[1] == "yml" {
		err := yaml.NewDecoder(file).Decode(conf)
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}
