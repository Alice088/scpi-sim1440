package device

import (
	"scpi-sim1440/internal/core"
	"scpi-sim1440/pkg/physic"
	"strconv"
	"sync"
	"time"
)

type BATConf struct {
	Name       string            `yaml:"name"`
	Capacity   physic.AmpereHour `yaml:"capacity"`
	EmptyLevel physic.Volt       `yaml:"empty_level"`
	FullLevel  physic.Volt       `yaml:"full_level"`
	Draw       physic.Ampere     `yaml:"draw"`
}

type BAT struct {
	name     string
	commands map[string]core.CommandHandler

	status     core.DeviceStatus
	charge     physic.AmpereHour
	capacity   physic.AmpereHour
	emptyLevel physic.Volt
	fullLevel  physic.Volt
	volt       physic.Volt
	draw       physic.Ampere
	mu         sync.Mutex
}

func NewBAT(conf BATConf) core.Device {
	device := BAT{
		status:     core.StatusIdle,
		name:       conf.Name,
		draw:       conf.Draw,
		capacity:   conf.Capacity,
		emptyLevel: conf.EmptyLevel,
		fullLevel:  conf.FullLevel,
		charge:     conf.Capacity,
	}

	device.commands = map[string]core.CommandHandler{
		"MEAS:CHARGE?": device.Charge,
		"MEAS:VOLT?":   device.Volt,
		"MEAS:CURR?":   device.Curr,
	}

	v, _ := strconv.ParseFloat(device.Volt(), 64)
	device.volt = v

	return &device
}

func (b *BAT) Status() core.DeviceStatus {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.status
}

func (b *BAT) Handle(cmd string) core.DeviceResponse {
	if res := defaultHandler(cmd, b.name, b.Reset); res != nil {
		return *res
	}

	if v, ok := b.commands[cmd]; ok {
		return core.DeviceResponse{
			Value: v(),
		}
	}

	return unknownHandler()
}

func (b *BAT) Tick(dt time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.charge -= b.draw * dt.Hours()
	if b.charge < 0 {
		b.charge = 0
	}
}

func (b *BAT) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = core.StatusIdle
	b.charge = b.capacity
}

func (b *BAT) Charge() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strconv.FormatFloat(b.charge, 'f', -1, 64)
}

func (b *BAT) Volt() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	soc := b.charge / b.capacity
	v := b.emptyLevel + soc*(b.fullLevel-b.emptyLevel)
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (b *BAT) Curr() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strconv.FormatFloat(b.draw, 'f', -1, 64)
}
