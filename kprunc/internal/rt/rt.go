package rt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/kprunc/internal/oci"
	"os"
)

type Runtime struct {
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func (rt *Runtime) Run(args []string) error {
	logger := logrus.New()
	file, err := os.OpenFile("/var/log/kprunc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer file.Close()
		logger.SetOutput(file)
	}

	lowLevelRuntime, err := oci.NewLowLevelRuntime(logger, []string{"runc"})
	if err != nil {
		return fmt.Errorf("could not create low level runtime: %v", err)
	}

	if !oci.HasCreateSubcommand(args) {
		logger.Infof("Skipping modifier for non-create subcommand")
		return lowLevelRuntime.Exec(args)
	}

	device, err := util.NewDevice()
	if err != nil {
		return lowLevelRuntime.Exec(args)
	}
	defer device.Shutdown()

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
