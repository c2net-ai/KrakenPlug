package labeler

import (
	"fmt"
	"golang.org/x/net/context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/cluster"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/utils"
	"strconv"
)

const (
	LabelKeyVendor = "krakenplug.pcl.ac.cn/device.vendor"
	LabelKeyMemory = "krakenplug.pcl.ac.cn/device.memory"
	LabelKeyModel  = "krakenplug.pcl.ac.cn/device.model"
)

type Labeler struct {
	client   *kubernetes.Clientset
	device   device.Device
	nodeName string
}

func NewLabeler(nodeName string) (*Labeler, error) {
	client, err := cluster.NewClient()
	if err != nil {
		return nil, err
	}
	device, err := util.NewDevice()

	if err != nil {
		return nil, err
	}
	return &Labeler{client: client, device: device, nodeName: nodeName}, nil
}

func (l *Labeler) Label() error {
	ctx := context.TODO()
	node, err := l.client.CoreV1().Nodes().Get(ctx, l.nodeName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %v", l.nodeName, err)
	}

	node.ObjectMeta.Labels[LabelKeyVendor] = l.device.Name()

	memoryInfo, err := l.device.GetDeviceMemoryInfo(0)
	if err != nil {
		klog.Errorf("failed to get device memory info: %v", err)
	}
	node.ObjectMeta.Labels[LabelKeyMemory] = strconv.Itoa(int(memoryInfo.Total))

	model, err := l.device.GetDeviceModel(0)
	if err != nil {
		klog.Errorf("failed to get device model: %v", err)
	}
	node.ObjectMeta.Labels[LabelKeyModel] = utils.ReplaceAllBlank(model)

	_, err = l.client.CoreV1().Nodes().Update(ctx, node, v1.UpdateOptions{})

	return err
}

func (l *Labeler) Shutdown() {
	if l.device != nil {
		l.device.Shutdown()
	}
}
