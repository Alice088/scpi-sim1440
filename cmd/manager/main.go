package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"scpi-sim1440/internal/parser"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const (
	actionUp   = "up"
	actionDown = "down"
	actionWork = "work"
)

const (
	imageName   = "standemu"
	networkName = "stand"
	port        = 5025
	subnet      = "172.30.0.0/24"
	baseIP      = "172.30.0.%d"
	firstIP     = 11
	busyboxHint = "To call a device, write: nc <ip> <port>"
)

type Device struct {
	Name string
	IP   string
	Args []string
}

func main() {
	rawAction := flag.String("action", actionUp, "up, down or work in docker")
	flag.Parse()

	if err := run(context.Background(), *rawAction); err != nil {
		log.Fatalf("%s: %v", *rawAction, err)
	}
}

func run(ctx context.Context, action string) error {
	switch action {
	case actionUp:
		return up(ctx)
	case actionDown:
		return down(ctx)
	case actionWork:
		return work(ctx)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func up(ctx context.Context) error {
	if err := buildImage(ctx); err != nil {
		return err
	}
	devices, err := loadDevices()
	if err != nil {
		return err
	}
	if err := runStand(ctx, devices); err != nil {
		return err
	}
	printSummary(actionUp, devices)
	return nil
}

func down(ctx context.Context) error {
	devices, err := loadDevices()
	if err != nil {
		return err
	}
	if err := stopStand(ctx, devices); err != nil {
		return err
	}
	printSummary(actionDown, devices)
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
	return cmd.Run()
}

func buildImage(ctx context.Context) error {
	log.Printf("building image %s", imageName)
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}
	return nil
}

func loadDevices() ([]Device, error) {
	parsed, err := parser.Parse("config/devices.yaml")
	if err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}

	devices := make([]Device, len(parsed))
	for i, dev := range parsed {
		devices[i] = Device{
			Name: dev.Name,
			IP:   fmt.Sprintf(baseIP, firstIP+i),
			Args: []string{"-devindex", strconv.Itoa(i)},
		}
	}
	return devices, nil
}

func printSummary(action string, devices []Device) {
	log.Printf("%s: %d devices", action, len(devices))
	for _, d := range devices {
		log.Printf("%s  %s:%d", d.Name, d.IP, port)
	}
}

func runStand(ctx context.Context, devices []Device) error {
	cli, err := newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := ensureNetwork(ctx, cli); err != nil {
		return err
	}
	for _, d := range devices {
		if err := ensureDevice(ctx, cli, d); err != nil {
			return err
		}
	}
	return nil
}

func stopStand(ctx context.Context, devices []Device) error {
	cli, err := newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, d := range devices {
		if err := cli.ContainerStop(ctx, d.Name, container.StopOptions{}); err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("stop %s: %w", d.Name, err)
		}
	}
	if err := cli.NetworkRemove(ctx, networkName); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("network remove: %w", err)
	}
	return nil
}

func ensureNetwork(ctx context.Context, cli *client.Client) error {
	_, err := cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{{Subnet: subnet}},
		},
	})
	if err != nil && !errdefs.IsConflict(err) {
		return fmt.Errorf("network create: %w", err)
	}
	return nil
}

func ensureDevice(ctx context.Context, cli *client.Client, d Device) error {
	insp, err := cli.ContainerInspect(ctx, d.Name)
	switch {
	case err == nil && insp.State.Running:
		log.Printf("%s already running, skip", d.Name)
		return verifyIP(ctx, cli, d)
	case err == nil:
		log.Printf("%s exists but stopped, recreating", d.Name)
		if err := cli.ContainerRemove(ctx, insp.ID, container.RemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("remove stale %s: %w", d.Name, err)
		}
	case !errdefs.IsNotFound(err):
		return fmt.Errorf("inspect %s: %w", d.Name, err)
	}

	if err := createAndStart(ctx, cli, d); err != nil {
		return err
	}
	return verifyIP(ctx, cli, d)
}

func createAndStart(ctx context.Context, cli *client.Client, d Device) error {
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{Image: imageName, Cmd: d.Args},
		&container.HostConfig{AutoRemove: true},
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

func newClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return cli, nil
}
