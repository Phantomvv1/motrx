package config

import (
	"errors"
	"slices"
	"sync"
)

type SelectionAlgorithm func() (*Server, error)

type Algorithms struct {
	config      *Config
	usedServers []*Server
	mu          sync.Mutex
}

func NewAlgorithms(config *Config) *Algorithms {
	return &Algorithms{
		config:      config,
		usedServers: nil,
		mu:          sync.Mutex{},
	}
}

func (a *Algorithms) RoundRobin() (*Server, error) {
	healthyServers := a.config.HealthyServers()
	if len(healthyServers) == 0 {
		return nil, errors.New("Error: there are no healthy servers")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, server := range healthyServers {
		if !slices.Contains(a.usedServers, server) {
			a.usedServers = append(a.usedServers, server)
			return server, nil
		}
	}

	server := healthyServers[0]
	a.usedServers = []*Server{server}

	return server, nil
}
