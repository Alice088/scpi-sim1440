package noise

import (
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"time"
)

type DevKind string

const (
	DevDelay   DevKind = "delay"
	DevGarbage DevKind = "garbage"
)

type DevEffect struct {
	Kind  DevKind
	Delay time.Duration
}

type ConnKind string

const (
	ConnSilence  ConnKind = "silence"
	ConnBreak    ConnKind = "break"
	ConnTruncate ConnKind = "truncate"
)

type ConnEffect struct {
	Kind  ConnKind
	CutAt int
}

// EffectPlan holds the effect chains for a single command.
// Device chain runs in fixed order: Delay (sleep), then Garbage (replace value), then Hold (suppress response).
// Conn chain runs in fixed order: Truncate (cut value), then Break (close after write), then Silence (suppress write).
// When the chain contains Hold or Silence, no response is written, but the other effects still run.
// Break and Truncate together: write the truncated value, then close.
type EffectPlan struct {
	Device []DevEffect
	Conn   []ConnEffect
}

type HistoryMoment struct {
	commandIndex int
	effect       EffectPlan
}

type Manager struct {
	conf      Conf
	history   []HistoryMoment
	cmdCount  int
	startedAt time.Time
}

func NewNoiseManager(conf Conf) Manager {
	return Manager{
		conf:      conf,
		startedAt: time.Now(),
	}
}

func (m *Manager) Plan() EffectPlan {
	m.cmdCount++
	var plan EffectPlan

	if m.triggered(m.conf.Delay.EffectConf) {
		plan.Device = append(plan.Device, DevEffect{Kind: DevDelay, Delay: m.conf.Delay.Value})
	}
	if m.triggered(m.conf.Garbage) {
		plan.Device = append(plan.Device, DevEffect{Kind: DevGarbage})
	}

	if m.triggered(m.conf.Truncate.EffectConf) {
		plan.Conn = append(plan.Conn, ConnEffect{Kind: ConnTruncate, CutAt: m.conf.Truncate.CutAt})
	}
	if m.triggered(m.conf.Break) {
		plan.Conn = append(plan.Conn, ConnEffect{Kind: ConnBreak})
	}
	if m.triggered(m.conf.Silence) {
		plan.Conn = append(plan.Conn, ConnEffect{Kind: ConnSilence})
	}

	if len(plan.Conn) > 0 || len(plan.Device) > 0 {
		m.history = append(m.history, HistoryMoment{
			commandIndex: m.cmdCount,
			effect:       plan,
		})
	}

	return plan
}

func (m *Manager) triggered(e EffectConf) bool {
	switch e.Mode {
	case EffectModeAlways:
		return true
	case EffectModeOnCommand:
		return m.cmdCount == e.N
	case EffectModeAfterTime:
		return time.Since(m.startedAt) > e.T
	case EffectModeChance:
		return m.roll(e.Name) < e.Chance
	}
	return false
}

func (m *Manager) roll(effect string) float64 {
	pack := fmt.Sprintf("%d:%d:%s", m.conf.Seed, m.cmdCount, effect)
	h := fnv.New64a()
	if _, err := fmt.Fprint(h, pack); err != nil {
		log.Printf("failed roll on %s\n", pack)
	}
	return float64(h.Sum64()) / float64(math.MaxUint64)
}
