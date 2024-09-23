package device

import (
	"fmt"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Device interface {
	Release() error
	GetDeviceCount() (int, error)
	GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error)
	IsDeviceHealthy(idx int) (bool, error)
	GetDeviceUtil(idx int) (float64, error)
	Name() string
	K8sResourceName() string
	GetDeviceMemoryUtil(idx int) (float64, error)
}

func K8sResourceName(name string) string {
	return fmt.Sprintf("krakenplug.pcl.ac.cn/%s", name)
}
