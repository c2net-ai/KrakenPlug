package util

// #cgo LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
// #include <dlfcn.h>
import "C"
import (
	"errors"
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/ascend"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/cambricon"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/enflame"
	"openi.pcl.ac.cn/c2net-ai/KrakenPlug/common/device/nvidia"
	"os"
	"path/filepath"
)

var (
	deviceLibs = map[string]string{
		device.Enflame:   "libefml.so.1.0.0",
		device.Cambricon: "libcndev.so",
		device.Ascend:    "libdcmi.so",
		device.Nvidia:    "libnvidia-ml.so.1",
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

func FindExecutableFile(file string) (bool, string) {
	dirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(dir, file))
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return true, matches[0]
		}
	}

	return false, ""
}
