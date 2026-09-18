package factory

import (
	"errors"
	"scpi-sim1440/internal/core"
	"scpi-sim1440/internal/device"
	"scpi-sim1440/internal/parser"
)

func DeviceFactory(rawd parser.RawDevice) (core.Device, error) {
	switch rawd.Type {
	case "bat":
		conf, err := rawd.BATConf()
		if err != nil {
			return nil, err
		}

		return device.NewBAT(conf), nil
	default:
		return nil, errors.New("unknown device: " + rawd.Type)
	}
}
