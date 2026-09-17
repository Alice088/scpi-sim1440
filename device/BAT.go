package device

import (
	"awesomeProject/core"
	"strconv"
	"sync"
	"time"
)

type BAT struct {
	name     string
	commands map[string]core.CommandHandler

	//todo убрать регистры и перейти на поля структуры
	register map[string]any
	mu       sync.Mutex
}

func NewBAT(name string) core.Device {
	device := BAT{
		name: name,
	}

	device.commands = map[string]core.CommandHandler{
		"MEAS:CHARGE?": device.charge,
		"MEAS:VOLT?":   device.volt,
		"MEAS:CURR?":   device.curr,
	}

	return &device
}

func (B *BAT) Status() core.DeviceStatus {
	B.mu.Lock()
	defer B.mu.Unlock()
	return B.register["status"].(core.DeviceStatus)
}

func (B *BAT) Handle(cmd string) core.DeviceResponse {
	if res := defaultHandler(cmd, B.name, B.Reset); res != nil {
		return *res
	}

	if v, ok := B.commands[cmd]; ok {
		return core.DeviceResponse{
			Value: v(),
		}
	}

	return unknownHandler()
}

func (B *BAT) Tick(duration time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (B *BAT) Reset() {
	B.mu.Lock()
	B.mu.Unlock()
	B.register["status"] = core.StatusIdle
}

func (B *BAT) charge() string {
	B.mu.Lock()
	defer B.mu.Unlock()
	return strconv.Itoa(B.register["charge"].(int))
}

func (B *BAT) volt() string {
	B.mu.Lock()
	defer B.mu.Unlock()
	return strconv.Itoa(B.register["volt"].(int))
}

func (B *BAT) curr() string {
	B.mu.Lock()
	defer B.mu.Unlock()
	return strconv.Itoa(B.register["curr"].(int))
}
