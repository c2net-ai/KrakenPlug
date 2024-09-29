# KrakenPlug

**KrakenPlug**是人工智能集群中管理异构AI计算设备的插件和工具集，其中异构设备插件实现AI设备注册Kubernetes、设备分配、健康状态上报等功能；异构设备发现插件将AI设备型号、个数等信息通过label的方式添加到Kubernetes Node中以供业务平台实现精细化管理；异构设备Prometheus Exporter可实时获取AI设备的利用率、显存、温度等运行指标。

## 安装部署

执行以下命令增加Chart仓库：

```
helm repo add krakenplug https://openi.pcl.ac.cn/Kraken/KrakenCharts/raw/branch/master
```

添加成功后同步仓库信息，如下：
```
helm repo update
```

这里可以先查看krakenplug已有的安装包版本，如下：
```
helm search repo krakenplug
```

然后将krakenplug对应版本的 chart 包下载到本地，并解压，如下：
```
helm pull krakenplug/krakenplug --version vx.x.x
tar -zxvf krakenplug-vx.x.x.tgz
```

### 配置values.yaml文件

部署krakenplug包时，需要修改部署包中的一些参数，这些参数都配置在安装包解压后目录中的values.yaml文件。

### 安装KrakenPlug
进入到解压后的krakenplug目录，执行helm install进行安装：

```
helm install krakenplug -n krakenplug ./  --values values.yaml
```

### 升级KrakenPlug
如果已经使用helm install成功安装过chart包，执行helm upgrade进行更新：

```
helm upgrade krakenplug -n krakenplug ./  --values values.yaml
```