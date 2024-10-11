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

type Device interface {
	Release() error
	GetDeviceCount() (int, error)
	GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error)
	IsDeviceHealthy(idx int) (bool, error)
	GetDeviceUtil(idx int) (int, error)
	Name() string
	K8sResourceName() string
	GetDeviceMemoryInfo(idx int) (*MemInfo, error)
}

func K8sResourceName(name string) string {
	return fmt.Sprintf("krakenplug.pcl.ac.cn/%s", name)
}
