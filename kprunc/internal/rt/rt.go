package rt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/kprunc/internal/oci"
)

type Runtime struct {
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

//func (rt *Runtime) Run(args []string) error {
//	runcPath, err := exec.LookPath("runc")
//	if err != nil {
//		return fmt.Errorf("could not find runc: %v", err)
//	}
//
//	runtimeArgs := []string{runcPath}
//	if len(args) > 1 {
//		runtimeArgs = append(runtimeArgs, args[1:]...)
//	}
//
//	oci.HasCreateSubcommand(args)
//
//	err = syscall.Exec(runtimeArgs[0], runtimeArgs, os.Environ())
//	if err != nil {
//		return fmt.Errorf("could not exec '%v': %v", args[0], err)
//	}
//	return nil
//}

func (rt *Runtime) Run(args []string) error {
	logger := logrus.New()
	lowLevelRuntime, err := oci.NewLowLevelRuntime(logger, []string{"runc"})
	if err != nil {
		return fmt.Errorf("could not create low level runtime: %v", err)
	}

	device, err := util.NewDevice()
	if !oci.HasCreateSubcommand(args) || err != nil {
		logger.Tracef("Skipping modifier for non-create subcommand")
		return lowLevelRuntime.Exec(args)
	}
	defer device.Shutdown()

	ociSpec, err := oci.NewSpec(logger, args)
	if err != nil {
		return fmt.Errorf("error constructing OCI specification: %v", err)
	}

	specModifier := NewModifySpec(device)

	return oci.NewModifyingRuntimeWrapper(
		logger,
		lowLevelRuntime,
		ociSpec,
		specModifier,
	).Exec(args)
}
