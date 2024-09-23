package cambricon

// #cgo LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include "cndev.h"
// #include <dlfcn.h>
import "C"
import (
	"fmt"
	"path/filepath"
	"time"
	"unsafe"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
)

type Cambricon struct {
	handles []unsafe.Pointer
}

func (c *Cambricon) GetDeviceMemoryUtil(idx int) (float64, error) {
	return 0, errors.New("not implement")
}

func (c *Cambricon) GetDeviceUtil(idx int) (float64, error) {
	return 0, errors.New("not implement")
}

func (c *Cambricon) IsDeviceHealthy(idx int) (bool, error) {
	var ret C.cndevRet_t
	var cardHealthState C.cndevCardHealthState_t
	var healthCode int
	cardHealthState.version = C.int(version)
	// sleep for some seconds
	time.Sleep(time.Duration(1) * time.Second)
	ret = C.cndevGetCardHealthState(&cardHealthState, C.int(idx))
	err := errorString(ret)
	if err != nil {
		return false, err
	}
	healthCode = int(cardHealthState.health)
	return !(healthCode == 0), nil
}

func (c *Cambricon) GetContainerAllocateResponse(idxs []int) (*pluginapi.ContainerAllocateResponse, error) {
	r := &pluginapi.ContainerAllocateResponse{}
	if hostDeviceExistsWithPrefix(mluMonitorDeviceName) {
		r.Devices = append(r.Devices, &pluginapi.DeviceSpec{
			HostPath:      mluMonitorDeviceName,
			ContainerPath: mluMonitorDeviceName,
			Permissions:   "rw",
		})
	}

	for i, id := range idxs {
		r.Devices = append(r.Devices, &pluginapi.DeviceSpec{
			HostPath:      fmt.Sprintf("%s%d", mluDeviceNamePrefix, id),
			ContainerPath: fmt.Sprintf("%s%d", mluDeviceNamePrefix, i),
			Permissions:   "rw",
		})
	}

	r.Mounts = append(r.Mounts, &pluginapi.Mount{
		ContainerPath: cnmonPath,
		HostPath:      cnmonPath,
		ReadOnly:      true,
	})

	return r, nil
}

func NewCambricon() (device.Device, error) {
	handle := C.dlopen(C.CString("libcndev.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == C.NULL {
		return nil, fmt.Errorf("load so failed")
	}
	r := C.cndevInit(C.int(0))
	err := errorString(r)
	if err != nil {
		return nil, err
	}

	c := &Cambricon{}
	c.handles = append(c.handles, handle)
	return c, nil
}

const (
	version              = 5
	mluDeviceNamePrefix  = "/dev/cambricon_dev"
	mluMonitorDeviceName = "/dev/cambricon_ctl"
	cnmonPath            = "/usr/bin/cnmon"
)

func errorString(cRet C.cndevRet_t) error {
	if cRet == C.CNDEV_SUCCESS {
		return nil
	}
	err := C.GoString(C.cndevGetErrorString(cRet))
	return fmt.Errorf("cndev: %v", err)
}

func (c *Cambricon) Release() error {
	ret := C.cndevRelease()
	if ret != C.CNDEV_SUCCESS {
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
	var cardInfos C.cndevCardInfo_t
	cardInfos.version = C.int(version)
	r := C.cndevGetDeviceCount(&cardInfos)
	return int(cardInfos.number), errorString(r)
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
	return "cambricon"
}

func (c *Cambricon) K8sResourceName() string {
	return device.K8sResourceName(c.Name())
}
