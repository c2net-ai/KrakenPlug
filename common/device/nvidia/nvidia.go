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
	nvmllib nvml.Interface
}

func (n *Nvidia) GetContainerVolume(idxs []int) *device.ContainerVolume {
	v := &device.ContainerVolume{}

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

	v.Binaries = []string{
		// utility_bins
		"nvidia-smi",          /* System management interface */
		"nvidia-debugdump",    /* GPU coredump utility */
		"nvidia-persistenced", /* Persistence mode utility */
		"nv-fabricmanager",    /* NVSwitch fabric manager utility */
		//"nvidia-modprobe",                /* Kernel module loader */
		//"nvidia-settings",                /* X server settings */
		//"nvidia-xconfig",                 /* X xorg.conf editor */

		// compute_bins
		"nvidia-cuda-mps-control", /* Multi process service CLI */
		"nvidia-cuda-mps-server",  /* Multi process service server */
		"kpsmi",
	}
	v.Libraries = []string{
		// utility_libs
		"libnvidia-ml.so",   /* Management library */
		"libnvidia-cfg.so",  /* GPU configuration */
		"libnvidia-nscq.so", /* Topology info for NVSwitches and GPUs */

		// compute_libs
		"libcuda.so",                   /* CUDA driver library */
		"libcudadebugger.so",           /* CUDA Debugger Library */
		"libnvidia-opencl.so",          /* NVIDIA OpenCL ICD */
		"libnvidia-gpucomp.so",         /* Shared Compiler Library */
		"libnvidia-ptxjitcompiler.so",  /* PTX-SASS JIT compiler (used by libcuda) */
		"libnvidia-fatbinaryloader.so", /* fatbin loader (used by libcuda) */
		"libnvidia-allocator.so",       /* NVIDIA allocator runtime library */
		"libnvidia-compiler.so",        /* NVVM-PTX compiler for OpenCL (used by libnvidia-opencl) */
		"libnvidia-pkcs11.so",          /* Encrypt/Decrypt library */
		"libnvidia-pkcs11-openssl3.so", /* Encrypt/Decrypt library (OpenSSL 3 support) */
		"libnvidia-nvvm.so",            /* The NVVM Compiler library */

		// video_libs
		"libvdpau_nvidia.so",       /* NVIDIA VDPAU ICD */
		"libnvidia-encode.so",      /* Video encoder */
		"libnvidia-opticalflow.so", /* NVIDIA Opticalflow library */
		"libnvcuvid.so",            /* Video decoder */

		// graphics_libs
		//"libnvidia-egl-wayland.so",       /* EGL wayland platform extension (used by libEGL_nvidia) */
		"libnvidia-eglcore.so", /* EGL core (used by libGLES*[_nvidia] and libEGL_nvidia) */
		"libnvidia-glcore.so",  /* OpenGL core (used by libGL or libGLX_nvidia) */
		"libnvidia-tls.so",     /* Thread local storage (used by libGL or libGLX_nvidia) */
		"libnvidia-glsi.so",    /* OpenGL system interaction (used by libEGL_nvidia) */
		"libnvidia-fbc.so",     /* Framebuffer capture */
		"libnvidia-ifr.so",     /* OpenGL framebuffer capture */
		"libnvidia-rtcore.so",  /* Optix */
		"libnvoptix.so",        /* Optix */

		// graphics_libs_glvnd
		//"libGLX.so",                      /* GLX ICD loader */
		//"libOpenGL.so",                   /* OpenGL ICD loader */
		//"libGLdispatch.so",               /* OpenGL dispatch (used by libOpenGL, libEGL and libGLES*) */
		"libGLX_nvidia.so",       /* OpenGL/GLX ICD */
		"libEGL_nvidia.so",       /* EGL ICD */
		"libGLESv2_nvidia.so",    /* OpenGL ES v2 ICD */
		"libGLESv1_CM_nvidia.so", /* OpenGL ES v1 common profile ICD */
		"libnvidia-glvkspirv.so", /* SPIR-V Lib for Vulkan */
		"libnvidia-cbl.so",       /* VK_NV_ray_tracing */

		// graphics_libs_compat
		"libGL.so",        /* OpenGL/GLX legacy _or_ compatibility wrapper (GLVND) */
		"libEGL.so",       /* EGL legacy _or_ ICD loader (GLVND) */
		"libGLESv1_CM.so", /* OpenGL ES v1 common profile legacy _or_ ICD loader (GLVND) */
		"libGLESv2.so",    /* OpenGL ES v2 legacy _or_ ICD loader (GLVND) */

		// ngx_libs
		"libnvidia-ngx.so", /* NGX library */

		// dxcore_libs
		"libdxcore.so", /* Core library for dxcore support */
	}
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
