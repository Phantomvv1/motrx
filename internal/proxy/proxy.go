package proxy

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Phantomvv1/motrx/internal/config"
)

func StartReverseProxy(config *config.Config) {
	go healthCheckServers(config)

	for {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			handleRequest(w, r, config)
		})

		log.Println("motrx is now listening on port 8000")

		err := http.ListenAndServe(":8000", mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func healthCheckServers(config *config.Config) {
	// The error values here are always nil since the main function checks if the config is valid
	interval, _ := config.GetInterval()
	timeout, _ := config.GetTimeout()
	http.DefaultClient.Timeout = timeout

	for {
		for _, server := range config.Servers {
			_, err := http.Get("http://" + server.Address + server.HealthCheckEndpoint)
			if err != nil {
				log.Printf("Error: %s is not responding, %v", server.Address, err)
				server.UpdateHealth(false)
			} else {
				server.UpdateHealth(true)
			}
		}

		time.Sleep(interval)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request, config *config.Config) {
	server, err := chooseServer(r, config)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	forwardRequest(server, w, r)
}

func chooseServer(r *http.Request, config *config.Config) (*config.Server, error) {
	alg, _ := config.Algorithm()
	server, err := alg()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func forwardRequest(server *config.Server, w http.ResponseWriter, r *http.Request) {
	target := "http://" + server.Address + r.URL.RequestURI()

	req, err := http.NewRequest(
		r.Method,
		target,
		r.Body,
	)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header.Clone()

	server.AddConnection()
	defer server.SubtractConnection()
	now := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	log.Printf("%s %s -> %s -> %d -> %v", r.Method, r.URL.Path, server.Address, resp.StatusCode, time.Since(now))
}
