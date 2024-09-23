package ascend

import (
	"context"
	"fmt"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/devmanager"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
)

type Ascend struct {
	dmgr devmanager.DeviceInterface
}

func (c *Ascend) GetDeviceMemoryUtil(idx int) (float64, error) {
	return 0, errors.New("not implement")
}

func (c *Ascend) GetDeviceUtil(idx int) (float64, error) {
	return 0, errors.New("not implement")
}

func (c *Ascend) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (c *Ascend) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["ASCEND_VISIBLE_DEVICES"] = idxsStr

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
	dmgr, err := devmanager.GetDeviceManager()
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

func (c *Ascend) Release() error {
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
