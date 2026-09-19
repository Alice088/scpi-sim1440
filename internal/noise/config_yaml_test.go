package noise

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestParseDur(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "", want: 0},
		{in: "150ms", want: 150 * time.Millisecond},
		{in: "2s", want: 2 * time.Second},
		{in: "nonsense", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseDur(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseDur(%q) error = nil, want error", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDur(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("parseDur(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestConfUnmarshalYAML(t *testing.T) {
	const doc = `seed: 1440
delay:
  mode: always
  value: 100ms
garbage:
  mode: chance
  chance: 0.05
silence:
  mode: on_command
  n: 30
break:
  mode: after_time
  t: 2s
truncate:
  mode: chance
  chance: 0.05
  cut_at: 5
`

	var c Conf
	if err := yaml.Unmarshal([]byte(doc), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if c.Seed != 1440 {
		t.Fatalf("seed = %d, want 1440", c.Seed)
	}
	if c.Delay.Mode != EffectModeAlways {
		t.Fatalf("delay mode = %q, want %q", c.Delay.Mode, EffectModeAlways)
	}
	if c.Delay.Value != 100*time.Millisecond {
		t.Fatalf("delay value = %v, want %v", c.Delay.Value, 100*time.Millisecond)
	}
	if c.Garbage.Mode != EffectModeChance || c.Garbage.Chance != 0.05 {
		t.Fatalf("garbage = %+v, want chance 0.05", c.Garbage)
	}
	if c.Silence.Mode != EffectModeOnCommand || c.Silence.N != 30 {
		t.Fatalf("silence = %+v, want on_command 30", c.Silence)
	}
	if c.Break.Mode != EffectModeAfterTime || c.Break.T != 2*time.Second {
		t.Fatalf("break = %+v, want after_time 2s", c.Break)
	}
	if c.Truncate.CutAt != 5 || c.Truncate.Chance != 0.05 {
		t.Fatalf("truncate = %+v, want cut_at 5 chance 0.05", c.Truncate)
	}
}

func TestConfUnmarshalInvalidDuration(t *testing.T) {
	var c Conf
	err := yaml.Unmarshal([]byte("delay:\n  value: bogus\n"), &c)
	if err == nil {
		t.Fatal("unmarshal error = nil, want error")
	}
	if !strings.Contains(err.Error(), "delay.value") {
		t.Fatalf("error = %q, want delay.value prefix", err.Error())
	}
}
