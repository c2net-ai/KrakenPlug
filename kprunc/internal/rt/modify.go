package rt

import (
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispecs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
)

type ModifySpec struct {
	device device.Device
}

func NewModifySpec(device device.Device) *ModifySpec {
	return &ModifySpec{device: device}
}

func (m *ModifySpec) Modify(spec *specs.Spec) error {
	response, err := m.device.GetContainerAllocateResponse([]int{0})
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
						Type:          "bind",
						Options:       []string{"rbind", "rprivate"},
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
