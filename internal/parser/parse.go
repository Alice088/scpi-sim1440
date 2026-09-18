package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Parse(path string) ([]RawDevice, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg.Devices, nil
}
