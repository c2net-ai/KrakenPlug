package enflame

import (
	"fmt"
	eflib "go-eflib"
	"go-eflib/efml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"math"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
)

type Enflame struct {
}

func (c *Enflame) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	_, total, used, err := eflib.GetDeviceMemoryInfo(efml.Handle{
		Dev_Idx: uint(idx),
	})
	if err != nil {
		return nil, err
	}

	return &device.MemInfo{
		Total: uint32(total),
		Used:  uint32(used),
	}, nil

}

func (c *Enflame) GetDeviceUtil(idx int) (int, error) {
	util, err := eflib.GetDeviceGcuUsage(efml.Handle{
		Dev_Idx: uint(idx),
	})
	if err != nil {
		return 0, err
	}

	return int(math.Round(float64(util))), nil
}

func (c *Enflame) IsDeviceHealthy(idx int) (bool, error) {
	return true, nil
}

func (c *Enflame) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["ENFLAME_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func NewEnflame() (device.Device, error) {
	err := eflib.Init(false)
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

func (c *Enflame) Release() error {
	eflib.Shutdown()
	return nil
}

func (c *Enflame) GetDeviceCount() (int, error) {
	count, err := eflib.GetDeviceCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get device count: %v", err)
	}

	return int(count), nil
}

func (c *Enflame) Name() string {
	return "enflame"
}

func (c *Enflame) K8sResourceName() string {
	return device.K8sResourceName(c.Name())
}
