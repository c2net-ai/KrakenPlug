package rt

import (
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispecs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"strconv"
	"strings"
)

type ModifySpec struct {
	device device.Device
}

const (
	EnvVisibleDevices = "KRAKENPLUG_VISIBLE_DEVICES"
)

func NewModifySpec(device device.Device) *ModifySpec {
	return &ModifySpec{device: device}
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

	var idxs []int
	split := strings.Split(devices, ",")
	for _, d := range split {
		i, err := strconv.Atoi(d)
		if err != nil {
			return nil
		}
		idxs = append(idxs, i)
	}

	if len(idxs) == 0 {
		return nil
	}

	response, err := m.device.GetContainerAllocateResponse(idxs)
	if err != nil {
		return nil
	}

	c := cdi.ContainerEdits{
		ContainerEdits: &cdispecs.ContainerEdits{},
	}

	for _, r := range response.Mounts {
		c.Append(&cdi.ContainerEdits{
			ContainerEdits: &cdispecs.ContainerEdits{
				Mounts: []*cdispecs.Mount{
					{
						ContainerPath: r.ContainerPath,
						HostPath:      r.HostPath,
						Options:       []string{"ro", "nosuid", "nodev", "bind"},
					},
				},
			},
		})
	}

	for _, d := range response.Devices {
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

	c.Apply(spec)

	return nil
}
