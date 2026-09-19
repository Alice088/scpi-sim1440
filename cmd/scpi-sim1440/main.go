package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/signal"
	"scpi-sim1440/internal/factory"
	"scpi-sim1440/internal/netx/handlers"
	"scpi-sim1440/internal/noise"
	"scpi-sim1440/internal/parser"
	"syscall"
	"time"
)

func main() {
	dindex := flag.Int("devindex", 0, "index of device that will up")
	flag.Parse()

	devices, err := parser.Parse("config/devices.yaml")
	if err != nil {
		log.Fatalf("failed parse devices: %v", err)
	}

	if len(devices) <= *dindex {
		log.Fatalf("out of range in devices; devices: %d but you get %d", len(devices), *dindex)
	}

	rawDev := devices[*dindex]

	ln, err := net.Listen("tcp", rawDev.Addr)
	if err != nil {
		log.Fatalf("failed listen %s on %s: %v", rawDev.Name, rawDev.Addr, err)
	}
	defer ln.Close()

	dev, err := factory.DeviceFactory(rawDev)
	if err != nil {
		log.Fatalf("failed get device from factory: %v", err)
	}

	visaHandler := handlers.NewVISAHandler(new(noise.NewNoiseManager(rawDev.Noise)), dev)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		t := time.Tick(time.Second)

		for {
			select {
			case <-t:
				dev.Tick(time.Second * time.Duration(rawDev.TimeScale))
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	log.Println("tcp up")
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("accept: %v", err)
				continue
			}
		}
		go visaHandler.Handle(conn)
	}
}
