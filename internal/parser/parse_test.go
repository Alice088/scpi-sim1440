package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validConfig = `devices:
  - name: bat-1
    type: bat
    addr: ":5025"
    time_scale: 60
    noise:
      seed: 1440
      delay:
        mode: always
        value: 100ms
      silence:
        mode: chance
        chance: 0.05
      break:
        mode: on_command
        n: 30
      truncate:
        mode: chance
        chance: 0.05
        cut_at: 5
    params:
      capacity: 2
      empty_level: 3
      full_level: 4.2
      draw: 1
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestParseValidConfig(t *testing.T) {
	devices, err := Parse(writeConfig(t, validConfig))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("devices len = %d, want 1", len(devices))
	}

	d := devices[0]
	if d.Name != "bat-1" {
		t.Fatalf("name = %q, want %q", d.Name, "bat-1")
	}
	if d.Type != "bat" {
		t.Fatalf("type = %q, want %q", d.Type, "bat")
	}
	if d.Addr != ":5025" {
		t.Fatalf("addr = %q, want %q", d.Addr, ":5025")
	}
	if d.TimeScale != 60 {
		t.Fatalf("time_scale = %v, want 60", d.TimeScale)
	}
	if d.Noise.Seed != 1440 {
		t.Fatalf("noise seed = %d, want 1440", d.Noise.Seed)
	}
	if d.Noise.Delay.Value != 100*time.Millisecond {
		t.Fatalf("noise delay value = %v, want %v", d.Noise.Delay.Value, 100*time.Millisecond)
	}
	if d.Noise.Truncate.CutAt != 5 {
		t.Fatalf("noise truncate cut_at = %d, want 5", d.Noise.Truncate.CutAt)
	}
	if d.Noise.Break.N != 30 {
		t.Fatalf("noise break n = %d, want 30", d.Noise.Break.N)
	}
}

func TestParseMissingFile(t *testing.T) {
	_, err := Parse(filepath.Join(t.TempDir(), "absent.yaml"))
	if err == nil {
		t.Fatal("Parse error = nil, want error")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Fatalf("error = %q, want read config prefix", err.Error())
	}
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse(writeConfig(t, "devices: ["))
	if err == nil {
		t.Fatal("Parse error = nil, want error")
	}
	if !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("error = %q, want parse config prefix", err.Error())
	}
}

func TestRawDeviceBATConf(t *testing.T) {
	devices, err := Parse(writeConfig(t, validConfig))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	conf, err := devices[0].BATConf()
	if err != nil {
		t.Fatalf("BATConf error: %v", err)
	}
	if conf.Name != "bat-1" {
		t.Fatalf("name = %q, want %q", conf.Name, "bat-1")
	}
	if conf.Capacity != 2 {
		t.Fatalf("capacity = %v, want 2", conf.Capacity)
	}
	if conf.EmptyLevel != 3 {
		t.Fatalf("empty_level = %v, want 3", conf.EmptyLevel)
	}
	if conf.FullLevel != 4.2 {
		t.Fatalf("full_level = %v, want 4.2", conf.FullLevel)
	}
	if conf.Draw != 1 {
		t.Fatalf("draw = %v, want 1", conf.Draw)
	}
}

func TestRawDeviceBATConfInvalidParams(t *testing.T) {
	doc := `devices:
  - name: bat-1
    type: bat
    params:
      capacity: [1, 2]
`
	devices, err := Parse(writeConfig(t, doc))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if _, err := devices[0].BATConf(); err == nil {
		t.Fatal("BATConf error = nil, want error")
	}
}
