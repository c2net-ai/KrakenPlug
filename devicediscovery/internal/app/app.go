package app

import (
	"github.com/urfave/cli"
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/devicediscovery/internal/labeler"
	"time"
)

const (
	ParamNodeName      = "node-name"
	ParamSleepInterval = "sleep-interval"
)

func NewApp(buildVersion ...string) *cli.App {
	c := cli.NewApp()
	c.Name = "Device Discovery"
	c.Usage = "Label the kubernetes node with device attributes related labels"
	if len(buildVersion) == 0 {
		buildVersion = append(buildVersion, "")
	}
	c.Version = buildVersion[0]

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:   ParamNodeName,
			Value:  "",
			Usage:  "Node name",
			EnvVar: "KRAKENPLUG_NODE_NAME",
		},
		&cli.DurationFlag{
			Name:   ParamSleepInterval,
			Value:  300 * time.Second,
			Usage:  "Time to sleep between labeling",
			EnvVar: "KRAKENPLUG_SLEEP_INTERVAL"},
	}

	c.Action = func(c *cli.Context) error {
		return action(c)
	}

	return c
}

func action(c *cli.Context) (err error) {
	config, err := contextToConfig(c)

	err = label(config.NodeName)
	if err != nil {
		klog.Infof("Failed to label node %s: %v", config.NodeName, err)
	}
	ticker := time.NewTicker(config.SleepInterval)
	defer ticker.Stop()
	for range ticker.C {
		err = label(config.NodeName)
		if err != nil {
			klog.Infof("Failed to label node %s: %v", config.NodeName, err)
		}
	}
	return nil
}

func label(nodeName string) error {
	labeler, err := labeler.NewLabeler(nodeName)
	defer func() {
		if labeler != nil {
			labeler.Shutdown()
		}
	}()
	if err != nil {
		return err
	}

	return labeler.Label()
}

func contextToConfig(c *cli.Context) (*Config, error) {
	return &Config{
		NodeName:      c.String(ParamNodeName),
		SleepInterval: c.Duration(ParamSleepInterval),
	}, nil
}
