package device

import "C"
import (
	"fmt"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	Enflame   = "enflame"
	Cambricon = "cambricon"
	Ascend    = "ascend"
	Nvidia    = "nvidia"
)

type MemInfo struct {
	Total uint32
	Used  uint32
}

type MountVolume struct {
	Binaries    []string `yaml:"binaries"`    // 可执行文件, 只需要填入文件名, 不需要带路径
	Libraries   []string `yaml:"libraries"`   // 动态库, 只需要填入文件名, 不需要带路径
	LibraryDirs []string `yaml:"libraryDirs"` // 动态库路径
}

type Device interface {
	Shutdown() error
	GetDeviceCount() (int, error)
	GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error)
	IsDeviceHealthy(idx int) (bool, error)
	GetDeviceUtil(idx int) (int, error)
	Name() string
	K8sResourceName() string
	GetDeviceMemoryInfo(idx int) (*MemInfo, error)
	GetDeviceModel(idx int) (string, error)
	GetDeviceVolume(idxs []int) []string
	GetMountVolume() *MountVolume
	SetMountVolumes(volume *MountVolume)
}

func K8sResourceName(name string) string {
	return fmt.Sprintf("krakenplug.pcl.ac.cn/%s", name)
}
