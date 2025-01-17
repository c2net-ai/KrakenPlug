package rt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/info"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/kprunc/internal/config"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/kprunc/internal/oci"
	"os"
	"strings"
)

type Runtime struct {
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func hasVersionFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		param := args[i]

		parts := strings.SplitN(param, "=", 2)
		trimmed := strings.TrimLeft(parts[0], "-")
		// If this is not a flag we continue
		if parts[0] == trimmed {
			continue
		}

		// Check the version flag
		if trimmed == "version" || trimmed == "v" {
			return true
		}
	}

	return false
}

func (rt *Runtime) GetConfig() (*config.Config, error) {
	file, err := ioutil.ReadFile("/etc/kprunc/config.yaml")
	if nil != err {
		return nil, fmt.Errorf("load config file failed: %v", err)
	}

	config := &config.Config{}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config file failed: %v", err)
	}

	return config, nil
}

func (rt *Runtime) Run(args []string) error {
	logger := logrus.New()
	file, err := os.OpenFile("/var/log/kprunc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer file.Close()
		logger.SetOutput(file)
	}

	printVersion := hasVersionFlag(args)
	if printVersion {
		fmt.Printf("%v version %v\n", "kprunc", info.GetVersionString())
	}

	lowLevelRuntime, err := oci.NewLowLevelRuntime(logger, []string{"runc"})
	if err != nil {
		return fmt.Errorf("could not create low level runtime: %v", err)
	}

	if !oci.HasCreateSubcommand(args) {
		logger.Infof("Skipping modifier for non-create subcommand")
		return lowLevelRuntime.Exec(args)
	}

	config, err := rt.GetConfig()
	if err != nil {
		return lowLevelRuntime.Exec(args)
	}

	device, err := util.NewDevice()
	if err != nil {
		return lowLevelRuntime.Exec(args)
	}
	defer device.Shutdown()
	device.SetMountVolumes(config.Volume[device.Name()])

	ociSpec, err := oci.NewSpec(logger, args)
	if err != nil {
		return fmt.Errorf("error constructing OCI specification: %v", err)
	}

	specModifier := NewModifySpec(logger, device)

	return oci.NewModifyingRuntimeWrapper(
		logger,
		lowLevelRuntime,
		ociSpec,
		specModifier,
	).Exec(args)
}
