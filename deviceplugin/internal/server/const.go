package server

import pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

const (
	ServerSock = pluginapi.DevicePluginPath + "openi.sock"
)
