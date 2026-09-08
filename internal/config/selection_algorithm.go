package config

import (
	"cmp"
	"errors"
	"math/rand"
	"slices"
	"sync"
)

type SelectionAlgorithm func() (*Server, error)

type Algorithms struct {
	config      *Config
	usedServers []*Server
	mu          sync.Mutex

	serverMap   map[*Server]int
	muServerMap sync.Mutex
}

func NewAlgorithms(config *Config) *Algorithms {
	return &Algorithms{
		config:      config,
		usedServers: nil,
		mu:          sync.Mutex{},
		serverMap:   make(map[*Server]int),
		muServerMap: sync.Mutex{},
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

func (a *Algorithms) Random() (*Server, error) {
	healthyServers := a.config.HealthyServers()
	if len(healthyServers) == 0 {
		return nil, errors.New("Error: there are no healthy servers")
	}

	num := rand.Intn(len(healthyServers))
	return healthyServers[num], nil
}

func (a *Algorithms) LeastConnections() (*Server, error) {
	healthyServers := a.config.HealthyServers()
	if len(healthyServers) == 0 {
		return nil, errors.New("Error: there are no healthy servers")
	}

	server := slices.MinFunc(healthyServers, func(serverA, serverB *Server) int {
		return cmp.Compare(serverA.Connections(), serverB.Connections())
	})

	return server, nil
}

func (a *Algorithms) WeightedRoundRobin() (*Server, error) {
	healthyServers := a.config.HealthyServers()
	if len(healthyServers) == 0 {
		return nil, errors.New("Error: there are no healthy servers")
	}

	a.muServerMap.Lock()
	defer a.muServerMap.Unlock()

	for _, server := range healthyServers {
		if a.serverMap[server] < server.Weight {
			a.serverMap[server] = a.serverMap[server] + 1
			return server, nil
		}
	}

	for i, server := range healthyServers {
		if i == 0 {
			a.serverMap[server] = 1
		} else {
			a.serverMap[server] = 0
		}
	}

	return healthyServers[0], nil
}
