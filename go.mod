module openi.pcl.ac.cn/Kraken/KrakenPlug

go 1.18

require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/prometheus/client_golang v1.15.0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli v1.22.5
	go-eflib v1.4.10
	google.golang.org/grpc v1.57.2
	huawei.com/npu-exporter/v6 v6.0.0-RC2.b002.fix
	k8s.io/klog/v2 v2.90.1
	k8s.io/kubelet v0.19.0
	k8s.io/kubernetes v1.14.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/tj/go-spin v1.1.0 // indirect
	github.com/xlab/c-for-go v1.3.0 // indirect
	github.com/xlab/pkgconfig v0.0.0-20170226114623-cea12a0fd245 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/cc/v4 v4.21.4 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sortutil v1.2.0 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
)

require (
	github.com/NVIDIA/go-nvml v0.12.4-0
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	k8s.io/apimachinery v0.19.0 // indirect
)

replace (
	go-eflib => ./thirdparty/go-eflib
	huawei.com/npu-exporter/v6 => gitee.com/lh120407/ascend-npu-exporter/v6 v6.0.0-RC2.b002.fix
)
