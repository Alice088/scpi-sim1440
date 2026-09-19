package noise

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type EffectConfRaw struct {
	Mode   string  `yaml:"mode"`
	N      int     `yaml:"n"`
	T      string  `yaml:"t"`
	Chance float64 `yaml:"chance"`
}

func (r EffectConfRaw) toConf() EffectConf {
	t, _ := time.ParseDuration(r.T)
	return EffectConf{Mode: r.Mode, N: r.N, T: t, Chance: r.Chance}
}

func parseDur(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

func (c *Conf) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Seed  int64 `yaml:"seed"`
		Delay struct {
			Mode   string  `yaml:"mode"`
			N      int     `yaml:"n"`
			T      string  `yaml:"t"`
			Chance float64 `yaml:"chance"`
			Value  string  `yaml:"value"`
		} `yaml:"delay"`
		Truncate struct {
			Mode   string  `yaml:"mode"`
			N      int     `yaml:"n"`
			T      string  `yaml:"t"`
			Chance float64 `yaml:"chance"`
			CutAt  int     `yaml:"cut_at"`
		} `yaml:"truncate"`
		Garbage EffectConfRaw `yaml:"garbage"`
		Silence EffectConfRaw `yaml:"silence"`
		Break   EffectConfRaw `yaml:"break"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}

	c.Seed = raw.Seed
	c.Garbage = raw.Garbage.toConf()
	c.Silence = raw.Silence.toConf()
	c.Break = raw.Break.toConf()

	c.Delay.EffectConf = EffectConf{Mode: raw.Delay.Mode, N: raw.Delay.N, Chance: raw.Delay.Chance}
	c.Truncate.EffectConf = EffectConf{Mode: raw.Truncate.Mode, N: raw.Truncate.N, Chance: raw.Truncate.Chance}
	c.Truncate.CutAt = raw.Truncate.CutAt

	var err error
	if c.Delay.Value, err = parseDur(raw.Delay.Value); err != nil {
		return fmt.Errorf("delay.value: %w", err)
	}
	if c.Delay.T, err = parseDur(raw.Delay.T); err != nil {
		return fmt.Errorf("delay.t: %w", err)
	}
	if c.Truncate.T, err = parseDur(raw.Truncate.T); err != nil {
		return fmt.Errorf("truncate.t: %w", err)
	}
	return nil
}
