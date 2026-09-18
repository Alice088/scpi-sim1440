package main

import (
	"log"
	"scpi-sim1440/internal/factory"
	"scpi-sim1440/internal/parser"
)

func main() {
	devices, err := parser.Parse("config/conf.yaml")
	if err != nil {
		log.Fatalf("failed parse con: %v", err)
	}

	for _, device := range devices {
		d, err := factory.DeviceFactory(device)
		if err != nil {
			log.Fatalf("failed get device from factory: %v", err)
		}

		log.Printf("%+v", d)
	}
}
