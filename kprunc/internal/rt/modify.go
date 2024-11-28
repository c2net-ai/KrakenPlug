package rt

import (
	"fmt"
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispecs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"huawei.com/npu-exporter/v6/common-utils/utils"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"os"
	"strconv"
	"strings"
)

type ModifySpec struct {
	device device.Device
	logger *logrus.Logger
}

const (
	EnvVisibleDevices = "KRAKENPLUG_VISIBLE_DEVICES"
)

var (
	mountOptions = []string{"ro", "nosuid", "nodev", "bind"}
)

func NewModifySpec(logger *logrus.Logger, device device.Device) *ModifySpec {
	return &ModifySpec{
		logger: logger,
		device: device,
	}
}

func (m *ModifySpec) Modify(spec *specs.Spec) error {
	envs := spec.Process.Env
	envDevices := ""
	for _, env := range envs {
		if strings.Contains(env, EnvVisibleDevices) {
			envDevices = env
			break
		}
	}

	devices := strings.Replace(envDevices, EnvVisibleDevices+"=", "", -1)
	if devices == "" {
		m.logger.Debugf("No device specified")
		return nil
	}

	var idxs []int
	if devices == "all" {
		cnt, err := m.device.GetDeviceCount()
		if err != nil {
			m.logger.Errorf("GetDeviceCount failed: %v", err)
			return nil
		}
		for i := 0; i < cnt; i++ {
			idxs = append(idxs, i)
		}
	} else {
		split := strings.Split(devices, ",")
		for _, d := range split {
			i, err := strconv.Atoi(d)
			if err != nil {
				m.logger.Errorf("Invalid device index: %v", d)
				return nil
			}
			idxs = append(idxs, i)
		}
	}

	if len(idxs) == 0 {
		return nil
	}

	m.logger.Infof("Device index: %v", idxs)
	response := m.device.GetContainerVolume(idxs)

	c := cdi.ContainerEdits{
		ContainerEdits: &cdispecs.ContainerEdits{},
	}

	for _, r := range response.Mounts {
		_, err := os.Stat(r.HostPath)
		if err != nil {
			m.logger.Errorf("Failed to find host path: %v", err)
			continue
		}
		c.Append(&cdi.ContainerEdits{
			ContainerEdits: &cdispecs.ContainerEdits{
				Mounts: []*cdispecs.Mount{
					{
						ContainerPath: r.ContainerPath,
						HostPath:      r.HostPath,
						Options:       mountOptions,
					},
				},
			},
		})
	}

	for _, d := range response.Devices {
		_, err := os.Stat(d.HostPath)
		if err != nil {
			m.logger.Errorf("Failed to find host path: %v", err)
			continue
		}
		m.logger.Infof("Device path: %v", d.HostPath)
		c.Append(&cdi.ContainerEdits{
			ContainerEdits: &cdispecs.ContainerEdits{
				DeviceNodes: []*cdispecs.DeviceNode{
					{
						HostPath: d.HostPath,
						Path:     d.HostPath,
					},
				},
			},
		})
	}

	for _, b := range response.Binaries {
		exist, path := util.FindExecutableFile(b)
		if exist {
			m.logger.Infof("Binary path: %v", path)
			c.Append(&cdi.ContainerEdits{
				ContainerEdits: &cdispecs.ContainerEdits{
					Mounts: []*cdispecs.Mount{
						{
							ContainerPath: fmt.Sprintf("/usr/local/bin/%v", b),
							HostPath:      path,
							Options:       mountOptions,
						},
					},
				},
			})
		}
	}

	for _, l := range response.Libraries {
		libPath, err := utils.GetDriverLibPath(l)
		if err != nil {
			continue
		}

		c.Append(&cdi.ContainerEdits{
			ContainerEdits: &cdispecs.ContainerEdits{
				Mounts: []*cdispecs.Mount{
					{
						ContainerPath: fmt.Sprintf("/usr/lib/%v", l),
						HostPath:      libPath,
						Options:       mountOptions,
					},
				},
			},
		})
	}

	c.Apply(spec)

	return nil
}
