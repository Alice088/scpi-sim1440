package noise

import "time"

type Conf struct {
	Delay         time.Duration
	DropChance    float64
	GarbageChance float64
	TimeScale     float64
}
