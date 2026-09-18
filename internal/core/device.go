package core

import "time"

type CommandHandler func() string

type Device interface {
	Status() DeviceStatus
	Handle(string) DeviceResponse
	Tick(time.Duration)
	Reset()
}

type DeviceResponse struct {
	Error error
	Value string
}
