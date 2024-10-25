package labeler

import (
	"golang.org/x/net/context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/cluster"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"
)

const (
	LabelKeyVendor = "krakenplug.pcl.ac.cn/device.vendor"
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
	defer func() {
		device.Release()
	}()
	if err != nil {
		return nil, err
	}
	return &Labeler{client: client, device: device, nodeName: nodeName}, nil
}

func (l *Labeler) Label() error {
	ctx := context.TODO()
	node, err := l.client.CoreV1().Nodes().Get(ctx, l.nodeName, v1.GetOptions{})
	if err != nil {
		return errors.Errorf(err, "failed to get node %s", l.nodeName)
	}

	node.ObjectMeta.Labels[LabelKeyVendor] = l.device.Name()

	_, err = l.client.CoreV1().Nodes().Update(ctx, node, v1.UpdateOptions{})

	return err
}
