package noise

import (
	"math/rand"
	"time"
)

type DevKind string

const (
	DevNone    DevKind = ""
	DevDelay   DevKind = "delay"
	DevGarbage DevKind = "garbage"
	DevHold    DevKind = "hold"
)

type DevEffect struct {
	Kind  DevKind
	Delay time.Duration
}

type ConnKind string

const (
	ConnNone     ConnKind = ""
	ConnBreak    ConnKind = "break"
	ConnTruncate ConnKind = "truncate"
)

type ConnEffect struct {
	Kind  ConnKind
	CutAt int
}

type EffectPlan struct {
	Device *DevEffect
	Conn   *ConnEffect
}

type HistoryMoment struct {
	commandIndex int
	effect       EffectPlan
}

type Manager struct {
	conf      Conf
	rng       *rand.Rand
	history   []HistoryMoment
	cmdCount  int
	startedAt time.Time
}

func (m *Manager) Plan() EffectPlan {
	m.cmdCount++
	var plan EffectPlan

	if m.triggered(m.conf.Delay.EffectConf) {
		plan.Device = &DevEffect{Kind: DevDelay, Delay: m.conf.Delay.Value}
	}
	if m.triggered(m.conf.Garbage) {
		plan.Device = &DevEffect{Kind: DevGarbage}
	}

	if m.triggered(m.conf.Break) {
		plan.Conn = &ConnEffect{Kind: ConnBreak}
	}
	if m.triggered(m.conf.Truncate.EffectConf) {
		plan.Conn = &ConnEffect{Kind: ConnTruncate, CutAt: m.conf.Truncate.CutAt}
	}

	if plan.Conn != nil || plan.Device != nil {
		m.history = append(m.history, HistoryMoment{
			commandIndex: m.cmdCount,
			effect:       plan,
		})
	}

	return plan
}

func (m *Manager) triggered(e EffectConf) bool {
	switch e.Mode {
	case "always":
		return true
	case "on_command":
		return m.cmdCount == e.N
	case "after_time":
		return time.Since(m.startedAt) > e.T
	}
	return false
}
