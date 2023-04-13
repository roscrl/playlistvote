package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Environment int

const (
	DEV Environment = iota
	PROD
)

// LoadEnv reads a .env file and sets environment variables
func LoadEnv(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			// Skip comments and empty lines
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}
