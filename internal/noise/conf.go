package noise

import "time"

type EffectConf struct {
	Mode   string
	N      int
	T      time.Duration
	Chance float64
}

type Conf struct {
	Seed     int64
	Delay    DelayConf
	Garbage  EffectConf
	Drop     EffectConf
	Break    EffectConf
	Truncate TruncateConf
}

type DelayConf struct {
	EffectConf
	Value time.Duration
}

type TruncateConf struct {
	EffectConf
	CutAt int
}
