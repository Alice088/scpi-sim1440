package handlers

import (
	"bufio"
	"net"
	"testing"
	"time"

	"scpi-sim1440/internal/device"
	"scpi-sim1440/internal/noise"
)

type visaConn struct {
	client net.Conn
	reader *bufio.Reader
}

func startVISA(t *testing.T, conf noise.Conf) *visaConn {
	t.Helper()

	mng := noise.NewNoiseManager(conf)
	dev := device.NewBAT(device.BATConf{
		Name:       "bat-1",
		Capacity:   2,
		EmptyLevel: 3,
		FullLevel:  4.2,
		Draw:       1,
	})
	h := NewVISAHandler(&mng, dev)

	server, client := net.Pipe()
	go h.Handle(server)
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})

	return &visaConn{client: client, reader: bufio.NewReader(client)}
}

func (v *visaConn) send(t *testing.T, cmd string) {
	t.Helper()
	if _, err := v.client.Write([]byte(cmd + "\n")); err != nil {
		t.Fatalf("write %q: %v", cmd, err)
	}
}

func (v *visaConn) read(t *testing.T) (string, error) {
	t.Helper()
	if err := v.client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	return v.reader.ReadString('\n')
}

func TestVISAHandlerRespondsToCommands(t *testing.T) {
	cases := []struct {
		cmd  string
		want string
	}{
		{cmd: "*IDN?", want: "bat-1\n"},
		{cmd: "MEAS:CHARGE?", want: "2.00\n"},
		{cmd: "MEAS:VOLT?", want: "4.20\n"},
		{cmd: "MEAS:CURR?", want: "1.00\n"},
	}

	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			v := startVISA(t, noise.Conf{})
			v.send(t, tc.cmd)
			got, err := v.read(t)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if got != tc.want {
				t.Fatalf("response = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVISAHandlerUnknownCommandReturnsEmptyLine(t *testing.T) {
	v := startVISA(t, noise.Conf{})
	v.send(t, "MEAS:TEMP?")

	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "\n" {
		t.Fatalf("response = %q, want %q", got, "\n")
	}
}

func TestVISAHandlerGarbageReplacesValue(t *testing.T) {
	v := startVISA(t, noise.Conf{
		Garbage: noise.EffectConf{Name: "garbage", Mode: noise.EffectModeAlways},
	})
	v.send(t, "MEAS:VOLT?")

	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "\ufffd#@%$\n" {
		t.Fatalf("response = %q, want garbage", got)
	}
}

func TestVISAHandlerTruncateCutsValue(t *testing.T) {
	v := startVISA(t, noise.Conf{
		Truncate: noise.TruncateConf{
			EffectConf: noise.EffectConf{Name: "truncate", Mode: noise.EffectModeAlways},
			CutAt:      2,
		},
	})
	v.send(t, "MEAS:VOLT?")

	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "4.\n" {
		t.Fatalf("response = %q, want %q", got, "4.\n")
	}
}

func TestVISAHandlerSilenceSuppressesResponse(t *testing.T) {
	v := startVISA(t, noise.Conf{
		Silence: noise.EffectConf{Name: "silence", Mode: noise.EffectModeAlways},
	})
	v.send(t, "MEAS:VOLT?")

	if err := v.client.SetReadDeadline(time.Now().Add(150 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	if got, err := v.reader.ReadString('\n'); err == nil {
		t.Fatalf("response = %q, want read timeout", got)
	}
}

func TestVISAHandlerBreakClosesAfterWrite(t *testing.T) {
	v := startVISA(t, noise.Conf{
		Break: noise.EffectConf{Name: "break", Mode: noise.EffectModeAlways},
	})
	v.send(t, "MEAS:VOLT?")

	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "4.20\n" {
		t.Fatalf("response = %q, want %q", got, "4.20\n")
	}

	if err := v.client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	if _, err := v.reader.ReadString('\n'); err == nil {
		t.Fatal("expected connection close after break")
	}
}

func TestVISAHandlerTruncateThenBreak(t *testing.T) {
	v := startVISA(t, noise.Conf{
		Truncate: noise.TruncateConf{
			EffectConf: noise.EffectConf{Name: "truncate", Mode: noise.EffectModeAlways},
			CutAt:      2,
		},
		Break: noise.EffectConf{Name: "break", Mode: noise.EffectModeAlways},
	})
	v.send(t, "MEAS:VOLT?")

	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "4.\n" {
		t.Fatalf("response = %q, want %q", got, "4.\n")
	}
}

func TestVISAHandlerDelayPostponesResponse(t *testing.T) {
	delay := 40 * time.Millisecond
	v := startVISA(t, noise.Conf{
		Delay: noise.DelayConf{
			EffectConf: noise.EffectConf{Name: "delay", Mode: noise.EffectModeAlways},
			Value:      delay,
		},
	})

	start := time.Now()
	v.send(t, "MEAS:VOLT?")
	got, err := v.read(t)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if elapsed := time.Since(start); elapsed < delay {
		t.Fatalf("elapsed = %v, want >= %v", elapsed, delay)
	}
	if got != "4.20\n" {
		t.Fatalf("response = %q, want %q", got, "4.20\n")
	}
}
