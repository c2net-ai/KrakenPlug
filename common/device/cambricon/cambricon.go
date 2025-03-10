package cambricon

// #cgo LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include <dlfcn.h>
import "C"
import (
	"fmt"
	"golang.org/x/sys/unix"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/cambricon/lib"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/utils"
	"path/filepath"
	"unsafe"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device"
)

const (
	version = 5
)

type Cambricon struct {
	handles     []unsafe.Pointer
	mountVolume *device.MountVolume
}

func (c *Cambricon) GetMountVolume() *device.MountVolume {
	return c.mountVolume
}

func (c *Cambricon) SetMountVolumes(volume *device.MountVolume) {
	c.mountVolume = volume
}

func (c *Cambricon) GetDeviceVolume(idxs []int) []string {
	v := []string{"/dev/cambricon_ctl"}

	for _, id := range idxs {
		v = append(v, fmt.Sprintf("%s%d", "/dev/cambricon_dev", id))
	}

	return v
}

func (c *Cambricon) GetDeviceModel(idx int) (string, error) {
	ret := lib.GetCardNameStringByDevId(int32(idx))
	if ret == nil {
		return "", fmt.Errorf("get card name failed, ret: %v", ret)
	}

	return unix.BytePtrToString(ret), nil
}

func (c *Cambricon) GetDeviceMemoryInfo(idx int) (*device.MemInfo, error) {
	memoryInfo := &lib.MemoryInfo_t{Version: version}
	ret := lib.GetMemoryUsage(memoryInfo, int32(idx))
	err := errorString(ret)
	if err != nil {
		return nil, err
	}

	return &device.MemInfo{
		Total: uint32(memoryInfo.PhysicalMemoryTotal),
		Used:  uint32(memoryInfo.PhysicalMemoryUsed),
	}, err
}

func (c *Cambricon) GetDeviceUtil(idx int) (int, error) {
	utilizationInfo := &lib.UtilizationInfo_t{
		Version: version,
	}
	ret := lib.GetDeviceUtilizationInfo(utilizationInfo, int32(idx))
	err := errorString(ret)
	if err != nil {
		return 0, err
	}

	return int(utilizationInfo.AverageCoreUtilization), err
}

func (c *Cambricon) IsDeviceHealthy(idx int) (bool, error) {
	//var ret lib.Ret_t
	//var cardHealthState lib.CardHealthState_t
	//var healthCode int
	//cardHealthState.Version = version
	//// sleep for some seconds
	//time.Sleep(time.Duration(1) * time.Second)
	//ret = lib.GetCardHealthState(&cardHealthState, int32(idx))
	//err := errorString(ret)
	//if err != nil {
	//	return false, err
	//}
	//healthCode = int(cardHealthState.Health)
	//return !(healthCode == 0), nil
	return true, nil
}

func (c *Cambricon) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}

	idxsStr := utils.JoinSliceInt(idxs)

	r.Envs = make(map[string]string)
	r.Envs["ASCEND_VISIBLE_DEVICES"] = idxsStr
	r.Envs["KRAKENPLUG_VISIBLE_DEVICES"] = idxsStr

	return r, nil
}

func NewCambricon() (device.Device, error) {
	handle := C.dlopen(C.CString("libcndev.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == C.NULL {
		return nil, fmt.Errorf("load so failed")
	}

	r := lib.Init(0)
	err := errorString(r)
	if err != nil {
		return nil, err
	}

	c := &Cambricon{}
	c.handles = append(c.handles, handle)
	return c, nil
}

func errorString(cRet lib.Ret_t) error {
	if cRet == lib.SUCCESS {
		return nil
	}
	err := lib.GetErrorString(cRet)
	return fmt.Errorf("cndev: %v", err)
}

func (c *Cambricon) Shutdown() error {
	ret := lib.Release()
	if ret != lib.SUCCESS {
		return errorString(ret)
	}

	for _, handle := range c.handles {
		err := C.dlclose(handle)
		if err != 0 {
			return fmt.Errorf("close handle failed")
		}
	}
	return nil
}

func (c *Cambricon) GetDeviceCount() (int, error) {
	cardInfos := &lib.CardInfo_t{
		Version: version,
	}
	r := lib.GetDeviceCount(cardInfos)
	return int(cardInfos.Number), errorString(r)
}

func hostDeviceExistsWithPrefix(prefix string) bool {
	matches, err := filepath.Glob(prefix + "*")
	if err != nil {
		klog.Infof("failed to know if host device with prefix exists, err: %v \n", err)
		return false
	}
	return len(matches) > 0
}

func (c *Cambricon) Name() string {
	return device.Cambricon
}

func (c *Cambricon) K8sResourceName() string {
	return device.K8sResourceName(c.Name())
}
