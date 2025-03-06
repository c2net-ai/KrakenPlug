package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/deviceplugin/internal/server"

	"k8s.io/klog/v2"

	"github.com/urfave/cli"
)

func NewApp(buildVersion ...string) *cli.App {
	c := cli.NewApp()
	c.Name = "Device Plugin"
	c.Usage = ""
	if len(buildVersion) == 0 {
		buildVersion = append(buildVersion, "")
	}
	c.Version = buildVersion[0]

	c.Flags = []cli.Flag{}

	c.Action = func(c *cli.Context) error {
		return action(c)
	}

	return c
}

func action(c *cli.Context) (err error) {
	klog.Info("Starting FS watcher.")
	watcher, err := startFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		return fmt.Errorf("created FS watcher: %v", err)
	}
	defer watcher.Close()

	klog.Info("Starting OS watcher.")
	sigs := startOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var devicePlugin *server.Server

restart:
	if devicePlugin != nil {
		devicePlugin.Stop()
	}
	devicePlugin, err = server.NewServer()
	if err != nil {
		klog.Info("Failed to create device plugin: " + err.Error())
		goto events
	}
	if err := devicePlugin.Serve(); err != nil {
		klog.Infof("serve device plugin err: %v, restarting.", err)
		goto events
	}

events:
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				klog.Infof("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				goto restart
			}
		case err := <-watcher.Errors:
			klog.Infof("inotify err: %v", err)
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				klog.Info("Received SIGHUP, restarting.")
				goto restart
			default:
				klog.Infof("Received signal %v, shutting down.", s)
				devicePlugin.Stop()
				break events
			}
		}
	}

	return nil
}

func startFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	return watcher, nil
}

func startOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}
