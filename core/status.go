package core

type DeviceStatus string

const (
	StatusIdle  DeviceStatus = "idle"
	StatusBusy  DeviceStatus = "busy"
	StatusFault DeviceStatus = "fault"
)
