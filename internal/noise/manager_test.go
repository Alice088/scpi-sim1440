package noise

import (
	"testing"
	"time"
)

func TestPlanCombinesDelayAndGarbageInDeviceChain(t *testing.T) {
	conf := Conf{
		Seed: 42,
		Delay: DelayConf{
			EffectConf: EffectConf{Name: "delay", Mode: EffectModeAlways},
			Value:      5 * time.Millisecond,
		},
		Garbage: EffectConf{Name: "garbage", Mode: EffectModeChance, Chance: 1.0},
	}
	m := NewNoiseManager(conf)

	plan := m.Plan()

	if len(plan.Device) != 2 {
		t.Fatalf("device chain len = %d, want 2", len(plan.Device))
	}
	if plan.Device[0].Kind != DevDelay {
		t.Fatalf("device[0] kind = %q, want %q", plan.Device[0].Kind, DevDelay)
	}
	if plan.Device[0].Delay != 5*time.Millisecond {
		t.Fatalf("device[0] delay = %v, want %v", plan.Device[0].Delay, 5*time.Millisecond)
	}
	if plan.Device[1].Kind != DevGarbage {
		t.Fatalf("device[1] kind = %q, want %q", plan.Device[1].Kind, DevGarbage)
	}
	if len(plan.Conn) != 0 {
		t.Fatalf("conn chain len = %d, want 0", len(plan.Conn))
	}
}

func TestPlanEmptyChainsAtZeroChance(t *testing.T) {
	conf := Conf{
		Seed:     7,
		Delay:    DelayConf{EffectConf: EffectConf{Name: "delay", Mode: EffectModeChance, Chance: 0.0}},
		Truncate: TruncateConf{EffectConf: EffectConf{Name: "truncate", Mode: EffectModeChance, Chance: 0.0}},
		Garbage:  EffectConf{Name: "garbage", Mode: EffectModeChance, Chance: 0.0},
		Silence:  EffectConf{Name: "silence", Mode: EffectModeChance, Chance: 0.0},
		Break:    EffectConf{Name: "break", Mode: EffectModeChance, Chance: 0.0},
	}
	m := NewNoiseManager(conf)

	plan := m.Plan()

	if len(plan.Device) != 0 {
		t.Fatalf("device chain len = %d, want 0", len(plan.Device))
	}
	if len(plan.Conn) != 0 {
		t.Fatalf("conn chain len = %d, want 0", len(plan.Conn))
	}
}
