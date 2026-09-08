package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintln(w, "Hello, World! I am server 1")
	})

	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintln(w, "Hello, World! I am server 2")
	})

	handler3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintln(w, "Hello, World! I am server 3")
	})

	go func() {
		log.Println("Server listening on :8001")
		if err := http.ListenAndServe(":8001", handler1); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Println("Server listening on :8002")
		if err := http.ListenAndServe(":8002", handler2); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Println("Server listening on :8003")
		if err := http.ListenAndServe(":8003", handler3); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}
