package enflame

import (
	"fmt"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"math"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/enflame/lib"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
)

type Enflame struct {
	mountVolume *device.MountVolume
}

func (c *Enflame) GetMountVolume() *device.MountVolume {
	return c.mountVolume
}

func (c *Enflame) SetMountVolumes(volume *device.MountVolume) {
	c.mountVolume = volume
}

func (c *Enflame) GetDeviceVolume(idxs []int) *device.DeviceVolume {
	return &device.DeviceVolume{}
}

func (c *Enflame) GetDeviceModel(idx int) (string, error) {
	return "", nil
}

func (c *Enflame) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	handle := lib.Handle{
		Dev_Idx: uint(idx),
	}
	memInfo, err := handle.GetDevMem()
	if err != nil {
		return nil, err
	}

	return &device.MemInfo{
		Total: uint32(memInfo.Mem_Total_Size / 1024 / 1024),
		Used:  uint32(memInfo.Mem_Used),
	}, nil

}

func (c *Enflame) GetDeviceUtil(idx int) (int, error) {
	handle := lib.Handle{
		Dev_Idx: uint(idx),
	}
	dtuUsage, err := handle.GetDevDtuUsage()
	if err != nil {
		return 0, err
	}

	return int(math.Round(float64(dtuUsage))), nil
}

func (c *Enflame) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (c *Enflame) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["ENFLAME_VISIBLE_DEVICES"] = idxsStr
	r.Envs["KRAKENPLUG_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func NewEnflame() (device.Device, error) {
	err := lib.InitV2(true)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize efml: %v", err)
	}

	c := &Enflame{}
	return c, nil
}

const (
	deviceNamePrefix = "/dev/gcu"
	deviceCtlPath    = "/dev/gcuctl"
	smiPath          = "/usr/sbin/efsmi"
)

func (c *Enflame) Shutdown() error {
	lib.Shutdown()
	return nil
}

func (c *Enflame) GetDeviceCount() (int, error) {
	count, err := lib.GetDevCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get device count: %v", err)
	}

	return int(count), nil
}

func (c *Enflame) Name() string {
	return device.Enflame
}

func (c *Enflame) K8sResourceName() string {
	return device.K8sResourceName(c.Name())
}
