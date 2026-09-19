package main

import (
	"log"
	"scpi-sim1440/internal/parser"
)

func main() {
	devices, err := parser.Parse("config/devices.yaml")
	if err != nil {
		log.Fatalf("failed parse devices: %v", err)
	}
	
}
