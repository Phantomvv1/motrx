package main

import (
	"log"
	"os"

	"github.com/Phantomvv1/motrx/internal/cli"
	"github.com/Phantomvv1/motrx/internal/config"
	"github.com/Phantomvv1/motrx/internal/proxy"
)

func main() {
	filePath, err := cli.StartCli(os.Args)
	if err != nil {
		log.Println(err)
		return
	}

	conf, err := config.ParseConfig(filePath)
	if err != nil {
		log.Println(err)
		return
	}

	if err = conf.Valid(); err != nil {
		log.Printf("Error: invalid config: %v", err)
		return
	}

	proxy.StartReverseProxy(conf)
}
