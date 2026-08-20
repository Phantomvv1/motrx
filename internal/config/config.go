package config

import (
	"log"
	"strings"
)

func ParseConfig(path string) {
	fileName := path
	index := strings.LastIndex(path, "/")
	if index != -1 {
		fileName = path[index+1:]
	}

	log.Println(fileName)
}
