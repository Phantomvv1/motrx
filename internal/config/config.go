package config

import "log"

func ParseConfig(path string) {
	fileName := ""
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			fileName = path[i+1:]
		}
	}

	log.Println(fileName)
}
