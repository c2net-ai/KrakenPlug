package rt

import (
	"fmt"
	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispecs "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/utils"
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
	if devices != "none" {
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
	}

	deviceVolume := m.device.GetDeviceVolume(idxs)
	mountVolume := m.device.GetMountVolume()

	c := cdi.ContainerEdits{
		ContainerEdits: &cdispecs.ContainerEdits{},
	}

	for _, d := range deviceVolume {
		_, err := os.Stat(d)
		if err != nil {
			m.logger.Errorf("Failed to find host path: %v", err)
			continue
		}
		c.Append(&cdi.ContainerEdits{
			ContainerEdits: &cdispecs.ContainerEdits{
				DeviceNodes: []*cdispecs.DeviceNode{
					{
						HostPath: d,
						Path:     d,
					},
				},
			},
		})
	}

	for _, b := range mountVolume.Binaries {
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

	for _, l := range mountVolume.Libraries {
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

	for _, dir := range mountVolume.LibraryDirs {
		if utils.IsExist(dir) {
			c.Append(&cdi.ContainerEdits{
				ContainerEdits: &cdispecs.ContainerEdits{
					Mounts: []*cdispecs.Mount{
						{
							ContainerPath: dir,
							HostPath:      dir,
							Options:       mountOptions,
						},
					},
				},
			})
		}
	}

	args := spec.Process.Args
	preCmd := "ldconfig > /dev/null 2>&1"

	if len(mountVolume.LibraryDirs) > 0 {
		preCmd = fmt.Sprintf(`echo '%s' >> /etc/ld.so.conf;%s`, strings.Join(mountVolume.LibraryDirs, "\n"), preCmd)
	}

	// 暂时先用这种方案，后续可考虑优化为hook时去ldconfig
	if isShellCmd(args) {
		spec.Process.Args = []string{args[0], args[1], fmt.Sprintf("%s;%s", preCmd, args[2])}
	} else {
		spec.Process.Args = []string{"sh", "-c", fmt.Sprintf("%s;exec %s", preCmd, strings.Join(args, " "))}
	}

	c.Apply(spec)
	return nil
}

func isShellCmd(args []string) bool {
	return len(args) == 3 &&
		(utils.StringInSlice(args[0], []string{"bash", "sh"}) || strings.Contains(args[0], "/bin/bash") || strings.Contains(args[0], "/bin/sh")) &&
		args[1] == "-c"
}
