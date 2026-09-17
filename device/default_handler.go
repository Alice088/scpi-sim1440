package device

import (
	"awesomeProject/core"
	"errors"
)

func defaultHandler(cmd, deviceName string, rst func()) *core.DeviceResponse {
	switch cmd {
	case "*RST":
		rst()
		return &core.DeviceResponse{
			Value: core.OK,
		}
	case "*IDN":
		return &core.DeviceResponse{
			Value: deviceName,
		}
	}

	return nil
}

func unknownHandler() core.DeviceResponse {
	return core.DeviceResponse{
		Error: errors.New("unknown command"),
	}
}
