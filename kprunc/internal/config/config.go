package config

import "openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device"

type Config struct {
	Volume map[string]*device.MountVolume `yaml:"volume"`
}
