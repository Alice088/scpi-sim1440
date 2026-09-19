package noise

import "time"

type EffectModeKind = string

const (
	EffectModeNone      EffectModeKind = ""
	EffectModeAlways    EffectModeKind = "always"
	EffectModeOnCommand EffectModeKind = "on_command"
	EffectModeAfterTime EffectModeKind = "after_time"
	EffectModeChance    EffectModeKind = "chance"
)

type EffectConf struct {
	Name   string
	Mode   string
	N      int
	T      time.Duration
	Chance float64
}

type Conf struct {
	Seed     int64
	Delay    DelayConf
	Truncate TruncateConf
	Garbage  EffectConf
	Silence  EffectConf
	Break    EffectConf
}

type DelayConf struct {
	EffectConf
	Value time.Duration
}

type TruncateConf struct {
	EffectConf
	CutAt int
}
