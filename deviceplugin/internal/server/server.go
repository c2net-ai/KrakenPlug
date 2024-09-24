package server

import (
	"context"
	"fmt"
	"net"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/nvidia"
	"os"
	"path"
	"strconv"
	"time"

	"k8s.io/klog/v2"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/ascend"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/cambricon"
	"openi.pcl.ac.cn/Kraken/KrakenPlug/common/device/enflame"

	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Server struct {
	devs   []*pluginapi.Device
	socket string
	stop   chan interface{}
	health chan *pluginapi.Device
	server *grpc.Server
	device device.Device
}

// NewServer returns an initialized Server
func NewServer() (*Server, error) {
	device, err := ascend.NewAscend()
	if err == nil {
		goto start
	}

	device, err = cambricon.NewCambricon()
	if err == nil {
		goto start
	}

	device, err = enflame.NewEnflame()
	if err == nil {
		goto start
	}

	device, err = nvidia.NewNvidia()
	if err == nil {
		goto start
	}

	return nil, err

start:
	s := &Server{
		socket: ServerSock,
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
		device: device,
	}

	count, err := device.GetDeviceCount()
	if err != nil {
		return nil, err
	}

	for i := 0; i < count; i++ {
		s.devs = append(s.devs, &pluginapi.Device{
			ID:     strconv.Itoa(i),
			Health: pluginapi.Healthy,
		})
	}

	return s, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *Server) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	go m.healthcheck()

	return nil
}

func (m *Server) healthcheck() {
	ctx, cancel := context.WithCancel(context.Background())
	health := make(chan *pluginapi.Device)

	go m.watchUnhealthy(ctx, health)

	for {
		select {
		case <-m.stop:
			cancel()
			return
		case dev := <-health:
			m.health <- dev
		}
	}
}

func (m *Server) watchUnhealthy(ctx context.Context, health chan<- *pluginapi.Device) {
	unhealthy := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, dm := range m.devs {
			idx, _ := strconv.Atoi(dm.ID)
			healthy, err := m.device.IsDeviceHealthy(idx)
			if err != nil {
				klog.Infof("Failed to get Device %s healthy status, set it as unhealthy", dm.ID)
				healthy = false
			}
			if !healthy && !unhealthy[dm.ID] {
				unhealthy[dm.ID] = true
				dev := pluginapi.Device{
					ID:     dm.ID,
					Health: pluginapi.Unhealthy,
				}
				health <- &dev
			} else if unhealthy[dm.ID] {
				delete(unhealthy, dm.ID)
				dev := pluginapi.Device{
					ID:     dm.ID,
					Health: pluginapi.Healthy,
				}
				health <- &dev
			}
		}

		//Sleep 1 second between two health checks
		time.Sleep(time.Second)
	}
}

// Stop stops the gRPC server
func (m *Server) Stop() error {
	if m.server == nil {
		return nil
	}

	m.server.Stop()
	m.server = nil
	close(m.stop)

	m.device.Release()
	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *Server) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
		Options:      &pluginapi.DevicePluginOptions{},
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (m *Server) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *Server) Serve() error {
	if err := m.Start(); err != nil {
		return fmt.Errorf("start device plugin err: %v", err)
	}

	klog.Infof("Starting to serve on socket %v", m.socket)
	resourceName := fmt.Sprintf("krakenplug.pcl.ac.cn/%s", m.device.Name())

	if err := m.Register(pluginapi.KubeletSocket, resourceName); err != nil {
		m.Stop()
		return fmt.Errorf("register resource %s err: %v", resourceName, err)
	}
	klog.Infof("Registered resource %s", resourceName)
	return nil
}

func (m *Server) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (m *Server) ListAndWatch(empty *pluginapi.Empty, server pluginapi.DevicePlugin_ListAndWatchServer) error {
	server.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			for i, dev := range m.devs {
				if dev.ID == d.ID {
					m.devs[i].Health = d.Health
					break
				}
			}
			server.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}

func (m *Server) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (m *Server) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	for _, req := range request.ContainerRequests {
		idxs := make([]int, 0)
		for _, id := range req.DevicesIDs {
			idx, _ := strconv.Atoi(id)
			idxs = append(idxs, idx)
		}
		car, err := m.device.GetContainerAllocateResponse(idxs)
		if err != nil {
			return nil, err
		}
		responses.ContainerResponses = append(responses.ContainerResponses, car)
	}
	return &responses, nil
}

func (m *Server) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
