package ascend

import (
	"context"
	"fmt"
	"huawei.com/npu-exporter/v6/devmanager/common"
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/devmanager"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
)

type Ascend struct {
	dmgr *devmanager.DeviceManager
}

func (c *Ascend) GetContainerVolume(idxs []int) *device.ContainerVolume {
	v := &device.ContainerVolume{}

	for i, id := range idxs {
		v.Devices = append(v.Devices, &device.DeviceSpec{
			HostPath:      fmt.Sprintf("/dev/davinci%d", id),
			ContainerPath: fmt.Sprintf("/dev/davinci%d", i),
		})
	}

	devManager := "/dev/davinci_manager"
	devSvm := "/dev/devmm_svm"
	devHdc := "/dev/hisi_hdc"
	v.Devices = append(v.Devices,
		&device.DeviceSpec{
			HostPath:      devManager,
			ContainerPath: devManager,
		},
		&device.DeviceSpec{
			HostPath:      devSvm,
			ContainerPath: devSvm,
		},
		&device.DeviceSpec{
			HostPath:      devHdc,
			ContainerPath: devHdc,
		},
	)

	v.Binaries = []string{
		"dcmi",
		"npu-smi",
	}

	v.LibraryDirs = []string{
		"/usr/local/Ascend/driver/lib64",
		"/usr/local/Ascend/driver/include",
	}

	return v
}

func (c *Ascend) GetDeviceModel(idx int) (string, error) {
	chipInfo, err := c.dmgr.GetChipInfo(int32(idx))
	if err != nil {
		return "", errors.Errorf(err, "failed to get product type")
	}

	return chipInfo.Name, nil
}

func (c *Ascend) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	hbmInfo, err := c.dmgr.GetDeviceHbmInfo(int32(idx))
	klog.Infof("memorySize: %d, usage: %d", hbmInfo.MemorySize, hbmInfo.Usage)
	if err != nil {
		return nil, errors.Errorf(nil, "failed to get device %d hbm info: %v", idx, err)
	}
	return &device.MemInfo{
		Total: uint32(hbmInfo.MemorySize),
		Used:  uint32(hbmInfo.Usage),
	}, nil
}

func (c *Ascend) GetDeviceUtil(idx int) (int, error) {
	rate, err := c.dmgr.GetDeviceUtilizationRate(int32(idx), common.AICore)
	if err != nil {
		return 0, errors.Errorf(nil, "failed to get device %d utilization rate: %v", idx, err)
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
