package util

// #cgo LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include <dlfcn.h>
import "C"
import (
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/ascend"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/cambricon"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/enflame"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/nvidia"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"
)

var (
	deviceLibs = map[string]string{
		device.Enflame:   "libefml.so",
		device.Cambricon: "libcndev.so",
		device.Ascend:    "libdcmi.so",
		device.Nvidia:    "libnvidia-ml.so",
	}
)

func NewDevice() (device.Device, error) {
	for deviceName, lib := range deviceLibs {
		handle := C.dlopen(C.CString(lib), C.RTLD_LAZY|C.RTLD_GLOBAL)
		if handle == C.NULL {
			continue
		}

		err := C.dlclose(handle)
		if err != 0 {
			return nil, errors.New("close handle failed")
		}

		switch deviceName {
		case device.Enflame:
			return enflame.NewEnflame()
		case device.Cambricon:
			return cambricon.NewCambricon()
		case device.Ascend:
			return ascend.NewAscend()
		case device.Nvidia:
			return nvidia.NewNvidia()
		}
	}

	klog.Info("not support device")

	return nil, errors.New("not support device")
}
