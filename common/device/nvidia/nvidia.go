package nvidia

import (
	"fmt"
	"k8s.io/klog/v2"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
)

type Nvidia struct {
	nvmllib nvml.Interface
}

func isSuccess(ret nvml.Return) bool {
	return ret == nvml.SUCCESS
}

func (n *Nvidia) Release() error {
	ret := n.nvmllib.Shutdown()
	if !isSuccess(ret) {
		klog.Infof("Error shutting down NVML: %v", ret)
		return errors.Errorf(nil, "Error shutting down NVML: %v", ret)
	}

	return nil
}

func (n *Nvidia) GetDeviceCount() (int, error) {
	count, ret := n.nvmllib.DeviceGetCount()
	if !isSuccess(ret) {
		return 0, errors.New("failed to get device count")
	}

	return count, nil
}

func (n *Nvidia) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["NVIDIA_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func (n *Nvidia) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (n *Nvidia) GetDeviceUtil(idx int) (float64, error) {
	return 0, nil
}

func (n *Nvidia) Name() string {
	return "nvidia"
}

func (n *Nvidia) K8sResourceName() string {
	return device.K8sResourceName(n.Name())
}

func (n *Nvidia) GetDeviceMemoryUtil(idx int) (float64, error) {
	device, r := n.nvmllib.DeviceGetHandleByIndex(idx)
	if !isSuccess(r) {
		return 0, errors.Errorf(nil, "Get DeviceHandleByIndex error")
	}
	memoryInfo, r := n.nvmllib.DeviceGetMemoryInfo(device)
	klog.Infof("Get MemoryInfo %v", memoryInfo)
	if !isSuccess(r) {
		return 0, errors.Errorf(nil, "Get MemoryInfo error")
	}
	if memoryInfo.Total == 0 {
		return 0, errors.Errorf(nil, "MemoryInfo total is 0")
	}

	return float64(memoryInfo.Used) / float64(memoryInfo.Total), nil
}

func NewNvidia() (device.Device, error) {
	nvmllib := nvml.New(
		nvml.WithLibraryPath("libnvidia-ml.so"),
	)
	ret := nvmllib.Init()
	if !isSuccess(ret) {
		return nil, fmt.Errorf("failed to initialize NVML: %v", ret)
	}

	return &Nvidia{nvmllib: nvmllib}, nil
}
