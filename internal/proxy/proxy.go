package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Phantomvv1/motrx/internal/config"
)

type RetryRequest struct {
	w            http.ResponseWriter
	r            *http.Request
	delay        time.Duration
	timesRetried int
	lastResp     *http.Response
	lastTS       time.Time
	server       *config.Server
}

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
	server, err := chooseServer(config)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	forwardRequest(config, server, w, r, config.Delay(), 0)
}

func chooseServer(config *config.Config) (*config.Server, error) {
	alg, _ := config.Algorithm()
	server, err := alg()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func forwardRequest(config *config.Config, server *config.Server, w http.ResponseWriter, r *http.Request, delay time.Duration, timesRetried int) {
	req, err := createReq(server, r)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	server.AddConnection()
	defer server.SubtractConnection()
	now := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		retryRequest(config, w, r, delay, timesRetried, nil, now, server)
		return
	}
	defer resp.Body.Close()

	if r.Method == http.MethodGet && resp.StatusCode/100 != 2 {
		retryRequest(config, w, r, delay, timesRetried, resp, now, server)
		return
	}

	writeResponse(w, resp, now, server, r)
}

func retryRequest(config *config.Config, w http.ResponseWriter, r *http.Request, delay time.Duration, timesRetried int, resp *http.Response, requestStartTS time.Time, server *config.Server) {
	if timesRetried >= config.Retry.MaxAttempts {
		writeResponse(w, resp, requestStartTS, server, r)
		return
	}

	time.Sleep(delay)

	server, err := chooseServer(config)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	forwardRequest(config, server, w, r, delay*2, timesRetried+1)
}

func createReq(server *config.Server, r *http.Request) (*http.Request, error) {
	target := "http://" + server.Address + r.URL.RequestURI()

	req, err := http.NewRequest(
		r.Method,
		target,
		r.Body,
	)
	if err != nil {
		return nil, errors.New("Error: failed to create request")
	}

	req.Header = r.Header.Clone()

	return req, nil
}

func writeResponse(w http.ResponseWriter, resp *http.Response, now time.Time, server *config.Server, r *http.Request) {
	if resp == nil {
		fmt.Fprint(w, "Unable to connect to a working server")
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if resp.Body == nil {
		log.Println("Bullshit")
	}

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	log.Printf("%s %s -> %s -> %d -> %v", r.Method, r.URL.Path, server.Address, resp.StatusCode, time.Since(now))
}
