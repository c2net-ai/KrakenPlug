package config

import "openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"

type Config struct {
	Volume map[string]*device.MountVolume `yaml:"volume"`
}
