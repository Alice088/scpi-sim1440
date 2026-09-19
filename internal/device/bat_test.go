package device

import (
	"testing"
	"time"

	"scpi-sim1440/internal/core"
)

func testBAT() core.Device {
	return NewBAT(BATConf{
		Name:       "bat-1",
		Capacity:   2,
		EmptyLevel: 3,
		FullLevel:  4.2,
		Draw:       1,
	})
}

func TestBATHandleCommands(t *testing.T) {
	cases := []struct {
		name    string
		cmd     string
		want    string
		wantErr bool
	}{
		{name: "idn", cmd: "*IDN?", want: "bat-1"},
		{name: "reset", cmd: "*RST", want: core.OK},
		{name: "charge", cmd: "MEAS:CHARGE?", want: "2.00"},
		{name: "volt", cmd: "MEAS:VOLT?", want: "4.20"},
		{name: "curr", cmd: "MEAS:CURR?", want: "1.00"},
		{name: "unknown", cmd: "MEAS:TEMP?", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dev := testBAT()
			resp := dev.Handle(tc.cmd)
			if tc.wantErr {
				if resp.Error == nil {
					t.Fatalf("Handle(%q) error = nil, want error", tc.cmd)
				}
				return
			}
			if resp.Error != nil {
				t.Fatalf("Handle(%q) unexpected error: %v", tc.cmd, resp.Error)
			}
			if resp.Value != tc.want {
				t.Fatalf("Handle(%q) = %q, want %q", tc.cmd, resp.Value, tc.want)
			}
		})
	}
}

func TestBATTickDischargesAndClamps(t *testing.T) {
	dev := testBAT()

	dev.Tick(time.Hour)
	if got := dev.Handle("MEAS:CHARGE?").Value; got != "1.00" {
		t.Fatalf("charge after 1h = %q, want %q", got, "1.00")
	}
	if got := dev.Handle("MEAS:VOLT?").Value; got != "3.60" {
		t.Fatalf("volt after 1h = %q, want %q", got, "3.60")
	}

	dev.Tick(5 * time.Hour)
	if got := dev.Handle("MEAS:CHARGE?").Value; got != "0.00" {
		t.Fatalf("charge after overdraw = %q, want %q", got, "0.00")
	}
	if got := dev.Handle("MEAS:VOLT?").Value; got != "3.00" {
		t.Fatalf("volt at empty = %q, want %q", got, "3.00")
	}
}

func TestBATResetRestoresCharge(t *testing.T) {
	dev := testBAT()

	dev.Tick(2 * time.Hour)
	resp := dev.Handle("*RST")
	if resp.Error != nil {
		t.Fatalf("reset error: %v", resp.Error)
	}
	if got := dev.Handle("MEAS:CHARGE?").Value; got != "2.00" {
		t.Fatalf("charge after reset = %q, want %q", got, "2.00")
	}
	if got := dev.Status(); got != core.StatusIdle {
		t.Fatalf("status after reset = %q, want %q", got, core.StatusIdle)
	}
}

func TestBATInitialStateIsFull(t *testing.T) {
	dev := testBAT()

	if got := dev.Status(); got != core.StatusIdle {
		t.Fatalf("initial status = %q, want %q", got, core.StatusIdle)
	}
	if got := dev.Handle("MEAS:CHARGE?").Value; got != "2.00" {
		t.Fatalf("initial charge = %q, want %q", got, "2.00")
	}
}
