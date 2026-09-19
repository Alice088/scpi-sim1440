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
	DevNone    DevKind = ""
	DevDelay   DevKind = "delay"
	DevGarbage DevKind = "garbage"
)

type DevEffect struct {
	Kind  DevKind
	Delay time.Duration
}

type ConnKind string

const (
	ConnNone     ConnKind = ""
	ConnSilence  ConnKind = "silence"
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
		plan.Device = &DevEffect{Kind: DevDelay, Delay: m.conf.Delay.Value}
	} else if m.triggered(m.conf.Garbage) {
		plan.Device = &DevEffect{Kind: DevGarbage}
	}

	if m.triggered(m.conf.Break) {
		plan.Conn = &ConnEffect{Kind: ConnBreak}
	} else if m.triggered(m.conf.Truncate.EffectConf) {
		plan.Conn = &ConnEffect{Kind: ConnTruncate, CutAt: m.conf.Truncate.CutAt}
	} else if m.triggered(m.conf.Silence) {
		plan.Conn = &ConnEffect{Kind: ConnSilence}
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
