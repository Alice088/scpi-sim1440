package parser

import (
	"scpi-sim1440/internal/device"
	"scpi-sim1440/internal/noise"

	"gopkg.in/yaml.v3"
)

type RawDevice struct {
	Name      string     `yaml:"name"`
	Type      string     `yaml:"type"`
	Addr      string     `yaml:"addr"`
	TimeScale float64    `yaml:"time_scale"`
	Noise     noise.Conf `yaml:"noise"`
	Params    yaml.Node  `yaml:"params"`
}

type Config struct {
	Devices []RawDevice `yaml:"devices"`
}

func (r RawDevice) BATConf() (device.BATConf, error) {
	var c device.BATConf
	if err := r.Params.Decode(&c); err != nil {
		return device.BATConf{}, err
	}
	c.Name = r.Name
	return c, nil
}
