package app

import (
	"net/http"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/signal"
	"os"
	"syscall"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"k8s.io/klog/v2"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/deviceexporter/internal/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/urfave/cli"
)

const (
	CLIAddress  = "address"
	CLINodeName = "node-name"
)

func NewApp(buildVersion ...string) *cli.App {
	c := cli.NewApp()
	c.Name = "Device Exporter"
	c.Usage = "Generates Device metrics in the prometheus format"
	if len(buildVersion) == 0 {
		buildVersion = append(buildVersion, "")
	}
	c.Version = buildVersion[0]

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:   CLIAddress,
			Value:  ":9400",
			Usage:  "Address",
			EnvVar: "KRAKENPLUG_NODE_Address",
		},
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
	if err != nil {
		return err
	}
	nodeName := config.NodeName
	if nodeName == "" {
		nodeName, err = os.Hostname()
		if err != nil {
			return errors.Errorf(err, "failed to get hostname")
		}
	}

	collectors := make([]prometheus.Collector, 0)
	collector, err := collector.NewCollector(nodeName)
	if err != nil {
		klog.Errorf("can not find any device, error: %v", err)
	} else {
		collectors = append(collectors, collector)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(collectors...)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	server := &http.Server{
		Addr: config.Address,
	}

	klog.Infof("start serving at %s", server.Addr)
	go server.ListenAndServe()

	sigs := signal.Signals(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		select {
		case s := <-sigs:
			switch s {
			default:
				collector.Shutdown()
				klog.Infof("shutdown")
			}
		}
	}
	return nil
}

func contextToConfig(c *cli.Context) (*collector.Config, error) {
	return &collector.Config{
		Address:  c.String(CLIAddress),
		NodeName: c.String(CLINodeName),
	}, nil
}
