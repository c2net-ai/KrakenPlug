package collector

import (
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/util"
	"strconv"
	"time"

	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/errors"

	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/deviceexporter/internal/podresources"
)

var (
	timeout = 10 * time.Second
	socket  = "/var/lib/kubelet/pod-resources/kubelet.sock"
	maxSize = 1024 * 1024 * 16 // 16 Mb
)

const (
	labelNodeName     = "node_name"
	labelPod          = "pod"
	labelDeviceIndex  = "device_index"
	labelVendor       = "vendor"
	metricDeviceUtil  = "krakenplug_device_util"
	metricMemoryUsed  = "krakenplug_memory_used"
	metricMemoryTotal = "krakenplug_memory_total"
)

type collector struct {
	device      device.Device
	client      podresources.PodResources
	metrics     map[string]metric
	nodeName    string
	deviceCount int
}

type labelValues struct {
	NodeName    string
	DeviceIndex string
	Vendor      string
	Pod         string
}

type metric struct {
	desc   *prometheus.Desc
	labels []string
}

func getMetric() map[string]metric {
	metrics := make(map[string]metric)
	utilLabels := []string{labelNodeName, labelPod, labelDeviceIndex, labelVendor}

	metrics[metricDeviceUtil] = metric{desc: prometheus.NewDesc(metricDeviceUtil, "device util", utilLabels, nil),
		labels: utilLabels}
	metrics[metricMemoryUsed] = metric{desc: prometheus.NewDesc(metricMemoryUsed, "memory free", utilLabels, nil),
		labels: utilLabels}
	metrics[metricMemoryTotal] = metric{desc: prometheus.NewDesc(metricMemoryTotal, "memory total", utilLabels, nil),
		labels: utilLabels}

	return metrics
}

func NewCollector(nodeName string) (prometheus.Collector, error) {
	device, err := util.NewDevice()
	if err != nil {
		return nil, errors.Errorf(err, "new device error")
	}

	klog.Infof("%s device found", device.Name())

	c := &collector{
		device:   device,
		client:   podresources.NewPodResourcesClient(timeout, socket, []string{device.K8sResourceName()}, maxSize),
		metrics:  getMetric(),
		nodeName: nodeName,
	}

	count, err := c.device.GetDeviceCount()
	if err != nil {
		return nil, errors.Errorf(err, "get device count error")
	}
	c.deviceCount = count
	return c, nil
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	podInfo, err := c.client.GetDeviceToPodInfo()
	if err != nil {
		klog.Errorf("GetDeviceToPodInfo error: %v", err)
	}
	klog.Infof("GetDeviceToPodInfo: %v", podInfo)
	c.collectDeviceUtil(ch, podInfo)
	c.collectDeviceMemory(ch, podInfo)
}

func (c *collector) getLabelValues(labels []string, values *labelValues) []string {
	labelValues := make([]string, 0)
	for _, label := range labels {
		switch label {
		case labelNodeName:
			labelValues = append(labelValues, values.NodeName)
		case labelPod:
			labelValues = append(labelValues, values.Pod)
		case labelDeviceIndex:
			labelValues = append(labelValues, values.DeviceIndex)
		case labelVendor:
			labelValues = append(labelValues, values.Vendor)
		}
	}

	return labelValues
}

func (c *collector) collectDeviceUtil(ch chan<- prometheus.Metric, podInfo map[string]podresources.PodInfo) {
	for i := 0; i < c.deviceCount; i++ {
		util, err := c.device.GetDeviceUtil(i)
		if err != nil {
			klog.Errorf("GetDeviceUtil error: %v", err)
			continue
		}

		values := &labelValues{}
		values.NodeName = c.nodeName
		values.DeviceIndex = strconv.Itoa(i)
		values.Vendor = c.device.Name()
		info, ok := podInfo[strconv.Itoa(i)]
		if ok {
			values.Pod = info.Pod
		}

		ch <- prometheus.MustNewConstMetric(c.metrics[metricDeviceUtil].desc, prometheus.GaugeValue, float64(util), c.getLabelValues(c.metrics[metricDeviceUtil].labels, values)...)
	}
}

func (c *collector) collectDeviceMemory(ch chan<- prometheus.Metric, podInfo map[string]podresources.PodInfo) {
	for i := 0; i < c.deviceCount; i++ {
		memoryInfo, err := c.device.GetDeviceMemoryInfo(i)
		if err != nil {
			klog.Errorf("GetDeviceUtil error: %v", err)
			continue
		}

		values := &labelValues{}
		values.NodeName = c.nodeName
		values.DeviceIndex = strconv.Itoa(i)
		values.Vendor = c.device.Name()
		info, ok := podInfo[strconv.Itoa(i)]
		if ok {
			values.Pod = info.Pod
		}

		ch <- prometheus.MustNewConstMetric(c.metrics[metricMemoryUsed].desc, prometheus.GaugeValue, float64(memoryInfo.Used), c.getLabelValues(c.metrics[metricMemoryUsed].labels, values)...)
		ch <- prometheus.MustNewConstMetric(c.metrics[metricMemoryTotal].desc, prometheus.GaugeValue, float64(memoryInfo.Total), c.getLabelValues(c.metrics[metricMemoryTotal].labels, values)...)

	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
}
