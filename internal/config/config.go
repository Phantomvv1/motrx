package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type Server struct {
	Address             string `json:"address"`
	HealthCheckEndpoint string `json:"health"`
}

type HealthCheck struct {
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type Retry struct {
	MaxAttempts int `json:"max_attempts"`
}

type Config struct {
	Servers                []Server    `json:"servers"`
	LoadBalancingAlgorithm string      `json:"load_balancing_algorithm"`
	HealthCheck            HealthCheck `json:"health_check"`
	Retry                  Retry       `json:"retry"`
}

func (c Config) GetInterval() (time.Duration, error) {
	if len(c.HealthCheck.Timeout) < 2 {
		return 0, errors.New("Error: invalid intevral duration")
	}

	switch c.HealthCheck.Interval[len(c.HealthCheck.Interval)-1] {
	case 's':
		if c.HealthCheck.Interval[len(c.HealthCheck.Interval)-2] == 'm' {
			interval := c.HealthCheck.Interval[0 : len(c.HealthCheck.Interval)-2]
			duration, err := strconv.Atoi(interval)
			if err != nil {
				return 0, err
			}

			return time.Millisecond * time.Duration(duration), nil
		}

		interval := c.HealthCheck.Interval[0 : len(c.HealthCheck.Interval)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Second * time.Duration(duration), nil
	case 'm':
		interval := c.HealthCheck.Interval[0 : len(c.HealthCheck.Interval)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Minute * time.Duration(duration), nil
	case 'h':
		interval := c.HealthCheck.Interval[0 : len(c.HealthCheck.Interval)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Hour * time.Duration(duration), nil
	default:
		return 0, errors.New("Error: invalid interval duration")
	}
}

func (c Config) GetTimeout() (time.Duration, error) {
	if len(c.HealthCheck.Timeout) < 2 {
		return 0, errors.New("Error: invalid timeout duration")
	}

	switch c.HealthCheck.Timeout[len(c.HealthCheck.Interval)-1] {
	case 's':
		if c.HealthCheck.Timeout[len(c.HealthCheck.Timeout)-2] == 'm' {
			interval := c.HealthCheck.Timeout[0 : len(c.HealthCheck.Timeout)-2]
			duration, err := strconv.Atoi(interval)
			if err != nil {
				return 0, err
			}

			return time.Millisecond * time.Duration(duration), nil
		}

		interval := c.HealthCheck.Timeout[0 : len(c.HealthCheck.Timeout)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Second * time.Duration(duration), nil
	case 'm':
		interval := c.HealthCheck.Timeout[0 : len(c.HealthCheck.Timeout)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Minute * time.Duration(duration), nil
	case 'h':
		interval := c.HealthCheck.Timeout[0 : len(c.HealthCheck.Timeout)-1]
		duration, err := strconv.Atoi(interval)
		if err != nil {
			return 0, err
		}

		return time.Hour * time.Duration(duration), nil
	default:
		return 0, errors.New("Error: invalid timeout duration")
	}
}

func (c Config) Valid() error {
	if _, err := c.GetInterval(); err != nil {
		return err
	}

	if _, err := c.GetTimeout(); err != nil {
		return err
	}

	for _, server := range c.Servers {
		if server.Address == "" || server.HealthCheckEndpoint == "" {
			return errors.New("Error: a server is missing an address or a health check endpoint")
		}
	}

	return nil
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
