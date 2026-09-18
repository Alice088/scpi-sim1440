package noise

import (
	"math/rand"
	"scpi-sim1440/internal/core"
	"time"
)

type Device struct {
	inner core.Device
	conf  Conf
	rng   *rand.Rand

	cmdCount  int
	startedAt time.Time
}

func NewNoiseDevice(d core.Device) core.Device {
	return Device{
		inner: d,
	}
}

func (d Device) Status() core.DeviceStatus {
	//TODO implement me
	panic("implement me")
}

func (d Device) Handle(s string) core.DeviceResponse {
	//TODO implement me
	panic("implement me")
}

func (d Device) Tick(duration time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (d Device) Reset() {
	//TODO implement me
	panic("implement me")
}
