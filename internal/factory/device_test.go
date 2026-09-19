package factory

import (
	"testing"

	"scpi-sim1440/internal/parser"

	"gopkg.in/yaml.v3"
)

func rawDevice(t *testing.T, doc string) parser.RawDevice {
	t.Helper()
	var cfg parser.Config
	if err := yaml.Unmarshal([]byte(doc), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if len(cfg.Devices) == 0 {
		t.Fatal("no devices in config")
	}
	return cfg.Devices[0]
}

func TestDeviceFactoryBuildsBAT(t *testing.T) {
	raw := rawDevice(t, `devices:
  - name: bat-1
    type: bat
    params:
      capacity: 2
      empty_level: 3
      full_level: 4.2
      draw: 1
`)

	dev, err := DeviceFactory(raw)
	if err != nil {
		t.Fatalf("DeviceFactory error: %v", err)
	}
	if dev == nil {
		t.Fatal("DeviceFactory returned nil device")
	}
	if got := dev.Handle("*IDN?").Value; got != "bat-1" {
		t.Fatalf("idn = %q, want %q", got, "bat-1")
	}
	if got := dev.Handle("MEAS:VOLT?").Value; got != "4.20" {
		t.Fatalf("volt = %q, want %q", got, "4.20")
	}
}

func TestDeviceFactoryUnknownType(t *testing.T) {
	raw := rawDevice(t, `devices:
  - name: osc-1
    type: osc
    params:
      foo: bar
`)

	dev, err := DeviceFactory(raw)
	if err == nil {
		t.Fatal("DeviceFactory error = nil, want error")
	}
	if dev != nil {
		t.Fatal("DeviceFactory returned device, want nil")
	}
}

func TestDeviceFactoryInvalidParams(t *testing.T) {
	raw := rawDevice(t, `devices:
  - name: bat-1
    type: bat
    params:
      capacity: not-a-number
`)

	if _, err := DeviceFactory(raw); err == nil {
		t.Fatal("DeviceFactory error = nil, want error")
	}
}
