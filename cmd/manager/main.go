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
	"time"

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
)

const (
	networkName = "stand"
	baseIP      = "172.30.0.%d"
	firstIP     = 11
)

func main() {
	rawAction := flag.String("action", "up", "up or down list of devices in docker")
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
			continue
		}
		if err == nil {
			log.Printf("%s exists but stopped, starting", d.Name)
			if err := cli.ContainerStart(ctx, insp.ID, container.StartOptions{}); err != nil {
				return fmt.Errorf("restart %s: %w", d.Name, err)
			}
			continue
		}

		resp, err := cli.ContainerCreate(ctx,
			&container.Config{
				Image: imageName,
				Cmd:   d.Args,
			},
			&container.HostConfig{AutoRemove: true},
			&network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					networkName: {IPAddress: d.IP},
				},
			},
			nil,
			d.Name,
		)
		if err != nil {
			return fmt.Errorf("create %s: %w", d.Name, err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return fmt.Errorf("start %s: %w", d.Name, err)
		}
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
		_ = cli.ContainerStop(ctx, d.Name, container.StopOptions{})
		if err := cli.ContainerRemove(ctx, d.Name, container.RemoveOptions{Force: true}); err != nil && !isNotFound(err) && !isInProgress(err) {
			return fmt.Errorf("remove %s: %w", d.Name, err)
		}
		if err := waitGone(ctx, cli, d.Name); err != nil {
			return fmt.Errorf("wait %s: %w", d.Name, err)
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

func isInProgress(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already in progress")
}

func waitGone(ctx context.Context, cli *client.Client, name string) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if _, err := cli.ContainerInspect(ctx, name); err != nil && isNotFound(err) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
