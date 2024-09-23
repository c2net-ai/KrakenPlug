module openi.pcl.ac.cn/Kraken/KrakenPlug/deviceexporter

go 1.18

require (
	github.com/prometheus/client_golang v1.15.0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli v1.22.5
	go.uber.org/automaxprocs v1.5.1
	google.golang.org/grpc v1.57.2
	k8s.io/klog/v2 v2.90.1
	k8s.io/kubernetes v1.14.1
	openi.pcl.ac.cn/Kraken/KrakenPlug/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20220104163920-15ed2e8cf2bd // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	go-eflib v1.4.10 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	huawei.com/npu-exporter/v6 v6.0.0-RC2.b002.fix // indirect
	k8s.io/apimachinery v0.26.2 // indirect
	k8s.io/kubelet v0.19.0 // indirect
)

replace (
	go-eflib => ../thirdparty/go-eflib
	huawei.com/npu-exporter/v6 => gitee.com/lh120407/ascend-npu-exporter/v6 v6.0.0-RC2.b002.fix
	openi.pcl.ac.cn/Kraken/KrakenPlug/common => ../common
)
