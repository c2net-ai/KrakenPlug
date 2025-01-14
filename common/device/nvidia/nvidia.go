package nvidia

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
)

type Nvidia struct {
	nvmllib     nvml.Interface
	mountVolume *device.MountVolume
}

func (n *Nvidia) GetMountVolume() *device.MountVolume {
	return n.mountVolume
}

func (n *Nvidia) SetMountVolumes(volume *device.MountVolume) {
	n.mountVolume = volume
}

func (n *Nvidia) GetDeviceVolume(idxs []int) *device.DeviceVolume {
	v := &device.DeviceVolume{}

	for i, id := range idxs {
		v.Devices = append(v.Devices, &device.DeviceSpec{
			HostPath:      fmt.Sprintf("/dev/nvidia%d", id),
			ContainerPath: fmt.Sprintf("/dev/nvidia%d", i),
		})
	}

	nvctl := "/dev/nvidiactl"
	nvuvm := "/dev/nvidia-uvm"
	nvuvmtools := "/dev/nvidia-uvm-tools"
	v.Devices = append(v.Devices,
		&device.DeviceSpec{
			HostPath:      nvctl,
			ContainerPath: nvctl,
		},
		&device.DeviceSpec{
			HostPath:      nvuvm,
			ContainerPath: nvuvm,
		},
		&device.DeviceSpec{
			HostPath:      nvuvmtools,
			ContainerPath: nvuvmtools,
		},
	)

	return v
}

func (n *Nvidia) GetDeviceModel(idx int) (string, error) {
	device, r := n.nvmllib.DeviceGetHandleByIndex(idx)
	if !isSuccess(r) {
		return "", fmt.Errorf("get device handle failed, ret: %v", r)
	}
	name, r := n.nvmllib.DeviceGetName(device)
	if !isSuccess(r) {
		return "", fmt.Errorf("get device name failed, ret: %v", r)
	}

	return name, nil
}

func (n *Nvidia) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	d, r := n.nvmllib.DeviceGetHandleByIndex(idx)
	if !isSuccess(r) {
		return nil, fmt.Errorf("get device handle failed, ret: %v", r)
	}
	memoryInfo, r := n.nvmllib.DeviceGetMemoryInfo_v2(d)
	if !isSuccess(r) {
		return nil, fmt.Errorf("get memery info failed, ret: %v", r)
	}

	return &device.MemInfo{
		Total: uint32(memoryInfo.Total / 1024 / 1024),
		Used:  uint32(memoryInfo.Used / 1024 / 1024)}, nil
}

func isSuccess(ret nvml.Return) bool {
	return ret == nvml.SUCCESS
}

func (n *Nvidia) Shutdown() error {
	ret := n.nvmllib.Shutdown()
	if !isSuccess(ret) {
		klog.Infof("shutting down nvml failed, ret: %v", ret)
		return fmt.Errorf("shutting down nvml failed, ret: %v", ret)
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
	r.Envs["KRAKENPLUG_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func (n *Nvidia) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (n *Nvidia) GetDeviceUtil(idx int) (int, error) {
	device, r := n.nvmllib.DeviceGetHandleByIndex(idx)
	if !isSuccess(r) {
		return 0, fmt.Errorf("get device handle failed, ret: %v", r)
	}
	util, r := n.nvmllib.DeviceGetUtilizationRates(device)
	if !isSuccess(r) {
		return 0, fmt.Errorf("get utilization rates failed, ret: %v", r)
	}

	return int(util.Gpu), nil
}

func (n *Nvidia) Name() string {
	return "nvidia"
}

func (n *Nvidia) K8sResourceName() string {
	return device.K8sResourceName(n.Name())
}

func NewNvidia() (device.Device, error) {
	nvmllib := nvml.New()
	ret := nvmllib.Init()
	if !isSuccess(ret) {
		return nil, fmt.Errorf("failed to initialize NVML: %v", ret)
	}

	return &Nvidia{nvmllib: nvmllib}, nil
}
