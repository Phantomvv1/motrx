package proxy

import (
	"log"
	"net/http"
	"time"

	"github.com/Phantomvv1/motrx/internal/config"
)

func StartReverseProxy(config *config.Config) {
	go healthCheckServers(config)
	for {
		time.Sleep(time.Second)
	}
}

func healthCheckServers(config *config.Config) {
	// The error values here are always nil since the main function checks if the config is valid
	interval, _ := config.GetInterval()
	// timeout, _ := config.GetTimeout()
	for {
		for _, server := range config.Servers {
			_, err := http.Get(server.HealthCheckEndpoint)
			if err != nil {
				log.Printf("Error: %s is not responding", server.Address)
			}
		}

		time.Sleep(interval)
	}
}
