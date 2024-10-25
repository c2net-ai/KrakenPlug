package app

import (
	"github.com/urfave/cli"
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/devicediscovery/internal/labeler"
	"time"
)

const (
	CLINodeName = "node-name"
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
			Name:   CLINodeName,
			Value:  "",
			Usage:  "Node name",
			EnvVar: "KRAKENPLUG_NODE_NAME",
		},
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
		klog.Infof("Failed to label node %s: %v", config.NodeName, errors.Message(err))
	}
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		err = label(config.NodeName)
		if err != nil {
			klog.Infof("Failed to label node %s: %v", config.NodeName, errors.Message(err))
		}
	}
	return nil
}

func label(nodeName string) error {
	labeler, err := labeler.NewLabeler(nodeName)
	defer labeler.Shutdown()
	if err != nil {
		return err
	}

	return labeler.Label()
}

func contextToConfig(c *cli.Context) (*Config, error) {
	return &Config{
		NodeName: c.String(CLINodeName),
	}, nil
}
