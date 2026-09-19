package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"scpi-sim1440/internal/parser"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const port = 5025
const imageName = "standemu"

type action = string

const (
	ActionUp   action = "up"
	ActionDown action = "down"
	ActionWork action = "work"
)

const busyboxHint = "To call a device, write: nc <ip> <port>"

const (
	networkName = "stand"
	baseIP      = "172.30.0.%d"
	firstIP     = 11
)

func main() {
	rawAction := flag.String("action", "up", "up, down or work in docker")
	flag.Parse()

	ctx := context.Background()

	switch action(*rawAction) {
	case ActionUp:
		if err := up(ctx); err != nil {
			log.Fatalf("up: %v", err)
		}
	case ActionDown:
		if err := down(ctx); err != nil {
			log.Fatalf("down: %v", err)
		}
	case ActionWork:
		if err := work(ctx); err != nil {
			log.Fatalf("work: %v", err)
		}
	default:
		log.Fatalf("unknown action: %s", *rawAction)
	}
}

func deviceList() ([]Device, error) {
	devices, err := parser.Parse("config/devices.yaml")
	if err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}

	list := make([]Device, 0, len(devices))
	for i, dev := range devices {
		list = append(list, Device{
			Name: dev.Name,
			IP:   fmt.Sprintf(baseIP, firstIP+i),
			Args: []string{"-devindex", strconv.Itoa(i)},
		})
	}
	return list, nil
}

func buildImage(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	log.Printf("image %s not found, building...", imageName)
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}
	return nil
}

func up(ctx context.Context) error {
	if err := buildImage(ctx); err != nil {
		return err
	}

	list, err := deviceList()
	if err != nil {
		return err
	}
	if err := RunStand(ctx, list); err != nil {
		return err
	}
	printSummary("up", list)
	return nil
}

func down(ctx context.Context) error {
	list, err := deviceList()
	if err != nil {
		return err
	}
	if err := StopStand(ctx, list); err != nil {
		return err
	}
	printSummary("down", list)
	return nil
}

func work(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "-it",
		"-e", "HINT="+busyboxHint,
		"--network", networkName,
		"busybox", "sh", "-c", `printf '%s\n' "$HINT"; exec sh`)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("busybox: %w", err)
	}
	return nil
}

func printSummary(verb string, list []Device) {
	log.Printf("%s: %d devices\n", verb, len(list))
	for _, d := range list {
		log.Printf("%s  %s:%d\n", d.Name, d.IP, port)
	}
}

type Device struct {
	Name string
	IP   string
	Args []string
}

func RunStand(ctx context.Context, devices []Device) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{{Subnet: "172.30.0.0/24"}},
		},
	})
	if err != nil && !isExists(err) {
		return fmt.Errorf("network create: %w", err)
	}

	for _, d := range devices {
		insp, err := cli.ContainerInspect(ctx, d.Name)
		if err == nil && insp.State.Running {
			log.Printf("%s already running, skip", d.Name)
			if err := verifyIP(ctx, cli, d); err != nil {
				return err
			}
			continue
		}
		if err == nil {
			log.Printf("%s exists but stopped, recreating", d.Name)
			if err := cli.ContainerRemove(ctx, insp.ID, container.RemoveOptions{Force: true}); err != nil && !isNotFound(err) {
				return fmt.Errorf("remove stale %s: %w", d.Name, err)
			}
		}

		resp, err := cli.ContainerCreate(ctx,
			&container.Config{
				Image: imageName,
				Cmd:   d.Args,
			},
			&container.HostConfig{
				AutoRemove: true,
			},
			nil,
			nil,
			d.Name,
		)
		if err != nil {
			return fmt.Errorf("create %s: %w", d.Name, err)
		}

		if err := cli.NetworkConnect(ctx, networkName, resp.ID, &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: d.IP},
		}); err != nil {
			return fmt.Errorf("network connect %s: %w", d.Name, err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return fmt.Errorf("start %s: %w", d.Name, err)
		}

		if err := verifyIP(ctx, cli, d); err != nil {
			return err
		}
	}
	return nil
}

func verifyIP(ctx context.Context, cli *client.Client, d Device) error {
	insp, err := cli.ContainerInspect(ctx, d.Name)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", d.Name, err)
	}
	got := ""
	if ep, ok := insp.NetworkSettings.Networks[networkName]; ok {
		got = ep.IPAddress
	}
	if got != d.IP {
		return fmt.Errorf("ip mismatch %s: want %s, got %s", d.Name, d.IP, got)
	}
	return nil
}

func StopStand(ctx context.Context, devices []Device) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, d := range devices {
		if err := cli.ContainerStop(ctx, d.Name, container.StopOptions{}); err != nil && !isNotFound(err) {
			return fmt.Errorf("stop %s: %w", d.Name, err)
		}
	}
	if err := cli.NetworkRemove(ctx, networkName); err != nil && !isNotFound(err) {
		return err
	}
	return nil
}

func isExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "No such container")
}
