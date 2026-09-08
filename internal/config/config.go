package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

type Server struct {
	Address             string `json:"address"`
	HealthCheckEndpoint string `json:"health"`
	Weight              int    `json:"weight"`
	healthy             bool
	mu                  sync.Mutex

	connections   int
	muConnections sync.Mutex
}

func (s *Server) UpdateHealth(status bool) {
	s.mu.Lock()
	s.healthy = status
	s.mu.Unlock()
}

func (s *Server) Healthy() bool {
	s.mu.Lock()
	health := s.healthy
	s.mu.Unlock()

	return health
}

func (s *Server) AddConnection() {
	s.muConnections.Lock()
	s.connections++
	s.muConnections.Unlock()
}

func (s *Server) SubtractConnection() {
	s.muConnections.Lock()
	s.connections--
	s.muConnections.Unlock()
}

func (s *Server) Connections() int {
	s.muConnections.Lock()
	conns := s.connections
	s.muConnections.Unlock()

	return conns
}

type HealthCheck struct {
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type Retry struct {
	MaxAttempts int `json:"max_attempts"`
}

type Config struct {
	Servers                []*Server   `json:"servers"`
	LoadBalancingAlgorithm string      `json:"load_balancing_algorithm"`
	HealthCheck            HealthCheck `json:"health_check"`
	Retry                  Retry       `json:"retry"`
	algorithms             *Algorithms
}

func (c *Config) setupConfig() {
	c.algorithms = NewAlgorithms(c)

	for _, server := range c.Servers {
		server.mu = sync.Mutex{}
		server.muConnections = sync.Mutex{}
	}

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
		if server.Address == "" {
			return errors.New("Error: a server is missing an address or a health check endpoint")
		}

		if server.Weight == 0 {
			server.Weight = 1
		}
	}

	_, ok := c.Algorithm()
	if !ok {
		log.Println("There wasn't an algorithm selected. Defaluting to Least Connections.")
	}

	return nil
}

func (c Config) HealthyServers() []*Server {
	healthyServers := []*Server{}
	for _, server := range c.Servers {
		if server.Healthy() {
			healthyServers = append(healthyServers, server)
		}
	}

	return healthyServers
}

func (c *Config) Algorithm() (SelectionAlgorithm, bool) {
	algorithms := map[string]SelectionAlgorithm{
		"round_robin":          c.algorithms.RoundRobin,
		"weighted_round_robin": c.algorithms.WeightedRoundRobin,
		"random":               c.algorithms.Random,
		"least_connections":    c.algorithms.LeastConnections,
	}

	if alg, ok := algorithms[c.LoadBalancingAlgorithm]; !ok {
		return c.algorithms.LeastConnections, ok
	} else {
		return alg, ok
	}
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

	conf.setupConfig()

	return conf, nil
}
