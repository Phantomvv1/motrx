package main

import (
	"log"
	"os"

	"github.com/Phantomvv1/motrx/internal/cli"
	"github.com/Phantomvv1/motrx/internal/config"
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

	log.Println(conf)
}
