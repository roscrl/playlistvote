package config

import (
	"fmt"
)

type Environment int

const (
	DEV Environment = iota
	PROD
)

func (e Environment) String() string {
	switch e {
	case DEV:
		return "DEV"
	case PROD:
		return "PROD"
	default:
		return "UNKNOWN"
	}
}

func parseEnvironment(value string) (Environment, error) {
	switch value {
	case "DEV":
		return DEV, nil
	case "PROD":
		return PROD, nil
	default:
		return 0, fmt.Errorf("unknown environment: %s", value)
	}
}
