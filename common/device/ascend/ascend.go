package ascend

import (
	"context"
	"fmt"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/devmanager"
	"huawei.com/npu-exporter/v6/devmanager/common"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/utils"
)

type Ascend struct {
	dmgr        *devmanager.DeviceManager
	mountVolume *device.MountVolume
}

func (c *Ascend) GetMountVolume() *device.MountVolume {
	return c.mountVolume
}

func (c *Ascend) SetMountVolumes(volume *device.MountVolume) {
	c.mountVolume = volume
}

func (c *Ascend) GetDeviceVolume(idxs []int) []string {
	v := []string{"/dev/davinci_manager", "/dev/devmm_svm", "/dev/hisi_hdc"}

	for _, id := range idxs {
		v = append(v, fmt.Sprintf("/dev/davinci%d", id))
	}

	return v
}

func (c *Ascend) GetDeviceModel(idx int) (string, error) {
	chipInfo, err := c.dmgr.GetChipInfo(int32(idx))
	if err != nil {
		return "", fmt.Errorf("get product type: %v", err)
	}

	return chipInfo.Name, nil
}

func (c *Ascend) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	hbmInfo, err := c.dmgr.GetDeviceHbmInfo(int32(idx))
	klog.Infof("memorySize: %d, usage: %d", hbmInfo.MemorySize, hbmInfo.Usage)
	if err != nil {
		return nil, fmt.Errorf("get device %d hbm info: %v", idx, err)
	}
	return &device.MemInfo{
		Total: uint32(hbmInfo.MemorySize),
		Used:  uint32(hbmInfo.Usage),
	}, nil
}

func (c *Ascend) GetDeviceUtil(idx int) (int, error) {
	rate, err := c.dmgr.GetDeviceUtilizationRate(int32(idx), common.AICore)
	if err != nil {
		return 0, fmt.Errorf("get device %d utilization rate: %v", idx, err)
	}
	return int(rate), nil
}

func (c *Ascend) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (c *Ascend) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["ASCEND_VISIBLE_DEVICES"] = idxsStr
	r.Envs["KRAKENPLUG_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func initHwLogger() error {
	var hwLogConfig = &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	return nil
}

func NewAscend() (device.Device, error) {
	initHwLogger()
	dmgr, err := devmanager.AutoInit("")
	if err != nil {
		return nil, err
	}
	err = dmgr.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ascend: %v", err)
	}

	c := &Ascend{dmgr: dmgr}
	return c, nil
}

func (c *Ascend) Shutdown() error {
	err := c.dmgr.ShutDown()
	if err != nil {
		return err
	}

	return nil
}

func (c *Ascend) GetDeviceCount() (int, error) {
	count, err := c.dmgr.GetDeviceCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get device count: %v", err)
	}
	return int(count), nil
}

func (c *Ascend) Name() string {
	return "ascend"
}

func (c *Ascend) K8sResourceName() string {
	return device.K8sResourceName(c.Name())
}
