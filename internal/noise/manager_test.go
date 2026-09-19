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

func TestPlanOnCommandTriggersOnlyAtN(t *testing.T) {
	conf := Conf{Seed: 1, Garbage: EffectConf{Name: "garbage", Mode: EffectModeOnCommand, N: 2}}
	m := NewNoiseManager(conf)

	if plan := m.Plan(); len(plan.Device) != 0 {
		t.Fatalf("command 1 device chain len = %d, want 0", len(plan.Device))
	}

	plan := m.Plan()
	if len(plan.Device) != 1 || plan.Device[0].Kind != DevGarbage {
		t.Fatalf("command 2 device chain = %+v, want single garbage", plan.Device)
	}

	if plan := m.Plan(); len(plan.Device) != 0 {
		t.Fatalf("command 3 device chain len = %d, want 0", len(plan.Device))
	}
}

func TestPlanChanceExtremes(t *testing.T) {
	always := NewNoiseManager(Conf{Garbage: EffectConf{Name: "garbage", Mode: EffectModeChance, Chance: 1}})
	for i := 0; i < 10; i++ {
		if plan := always.Plan(); len(plan.Device) != 1 {
			t.Fatalf("command %d chance=1 device chain len = %d, want 1", i+1, len(plan.Device))
		}
	}

	never := NewNoiseManager(Conf{Garbage: EffectConf{Name: "garbage", Mode: EffectModeChance, Chance: 0}})
	for i := 0; i < 10; i++ {
		if plan := never.Plan(); len(plan.Device) != 0 {
			t.Fatalf("command %d chance=0 device chain len = %d, want 0", i+1, len(plan.Device))
		}
	}
}

func TestPlanAfterTimeTriggers(t *testing.T) {
	conf := Conf{Break: EffectConf{Name: "break", Mode: EffectModeAfterTime, T: -time.Second}}
	m := NewNoiseManager(conf)

	plan := m.Plan()
	if len(plan.Conn) != 1 || plan.Conn[0].Kind != ConnBreak {
		t.Fatalf("conn chain = %+v, want single break", plan.Conn)
	}
}

func TestPlanRecordsHistoryOnlyWhenTriggered(t *testing.T) {
	conf := Conf{Seed: 5, Break: EffectConf{Name: "break", Mode: EffectModeOnCommand, N: 2}}
	m := NewNoiseManager(conf)

	m.Plan()
	if len(m.history) != 0 {
		t.Fatalf("history len after non-effecting command = %d, want 0", len(m.history))
	}

	m.Plan()
	if len(m.history) != 1 {
		t.Fatalf("history len after effecting command = %d, want 1", len(m.history))
	}
	if m.history[0].commandIndex != 2 {
		t.Fatalf("history command index = %d, want 2", m.history[0].commandIndex)
	}
	if m.cmdCount != 2 {
		t.Fatalf("cmdCount = %d, want 2", m.cmdCount)
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
