package rt

import (
	"fmt"
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispecs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
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
			c.Append(&cdi.ContainerEdits{
				ContainerEdits: &cdispecs.ContainerEdits{
					Mounts: []*cdispecs.Mount{
						{
							ContainerPath: fmt.Sprintf("/usr/local/bin/%v", b),
							//ContainerPath: path,
							HostPath: path,
							Options:  mountOptions,
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
						//ContainerPath: libPath,
						ContainerPath: fmt.Sprintf("/usr/lib/%v", libPath[strings.LastIndex(libPath, "/"):]),
						HostPath:      libPath,
						Options:       mountOptions,
					},
				},
			},
		})
	}

	args := spec.Process.Args
	ldconfig := "ldconfig > /dev/null 2>&1;"

	//shell方式的启动命令这样修改没影响，但是如果exec改为shell启动会影响终止等信号传递，以及环境变量行为上有些差异。
	//暂时先用这种方案，后续可考虑优化为hook时去ldconfig。
	if len(args) == 3 && (utils.StringInSlice(args[0], []string{"bash", "sh"})) && args[1] == "-c" {
		spec.Process.Args = []string{args[0], args[1], fmt.Sprintf("%s%s", ldconfig, args[2])}
	}

	c.Apply(spec)
	return nil
}
