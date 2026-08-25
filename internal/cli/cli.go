package cli

import (
	"errors"
)

func StartCli(args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("Error: there are not enough arguments provided")
	}

	if args[1] != "--config" {
		return "", errors.New("Error: there is no config specified")
	}

	return args[2], nil
}
