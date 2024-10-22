// ///////////////////////////////////////////////////////////////////////////
//
//	@brief API interface of Enflame Managerment Library
//	Efml API Binding for Go
//	Enflame Tech, All Rights Reserved. 2023 Copyright (C)
//
// ///////////////////////////////////////////////////////////////////////////
package lib

// #cgo LDFLAGS: -ldl  -Wl,--unresolved-symbols=ignore-in-object-files
// #include "stdbool.h"
// #include "efml.h"
import "C"

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type Handle struct {
	Dev_Idx uint
}

const (
	szName = C.MAX_CHAR_BUFF_LEN
	szUUID = C.MAX_CHAR_BUFF_LEN
)

type DevThermalInfo struct {
	Cur_Dev_Temp  float32
	Cur_Hbm0_Temp float32
	Cur_Hbm1_Temp float32
}

type DevPowerInfo struct {
	Pwr_Capability      float32
	Cur_Pwr_Consumption float32
}

type DevMemInfo struct {
	Mem_Total_Size uint
	Mem_Used       uint
}

type ClusterHbmMemInfo struct {
	Mem_Total_Size uint
	Mem_Used       uint
}

type DevClkInfo struct {
	Cur_Hbm_Clock uint
	Cur_Dtu_Clock uint
}

type DeviceInfo struct {
	Name      string
	Vendor_Id uint
	Device_Id uint
	Domain_Id uint
	Bus_Id    uint
	Dev_Id    uint
	Func_Id   uint
}

type EventInfo struct {
	Id   uint
	Type uint
	Msg  string
}

type LinkInfo struct {
	Link_Speed     uint
	Max_Link_Speed uint
	Link_Width     uint
	Max_Link_Width uint
}

type ThroughputInfo struct {
	Tx_Throughput float32
	Rx_Throughput float32
	Tx_Nak        uint64
	Rx_Nak        uint64
}

type EslPortInfo struct {
	Connected        uint
	Uuid             string
	Vendor_Id        uint
	Device_Id        uint
	Domain_Id        uint
	Bus_Id           uint
	Dev_Id           uint
	Func_Id          uint
	Port_Id          uint
	Port_Type        uint
	Remote_Card_Id   uint
	Remote_Uuid      string
	Remote_Vendor_Id uint
	Remote_Device_Id uint
	Remote_Domain_Id uint
	Remote_Bus_Id    uint
	Remote_Dev_Id    uint
	Remote_Func_Id   uint
	Remote_Port_Id   uint
	Remote_Port_Type uint
}

type DevEccStatus struct {
	Enabled bool
	Pending bool
	Pdblack bool
	Ecnt_sb uint
	Ecnt_db uint
}

type ProcessInfo struct {
	Pid         uint
	DevMemUsage uint64
	SysMemUsage uint64
}

type DevRmaStatus struct {
	SupportRma bool
	Flags      bool
}

type DevRmaDetails struct {
	SupportRma bool
	Flags      bool
	Dbe        uint
}

// utils function
func uintPtr(c C.uint) *uint {
	i := uint(c)
	return &i
}

func uint64Ptr(c C.uint64_t) *uint64 {
	i := uint64(c)
	return &i
}

func floatPtr(c C.float) *float32 {
	i := float32(c)
	return &i
}

func stringPtr(c *C.char) *string {
	s := C.GoString(c)
	return &s
}

func errorString(ret C.efmlReturn_t) error {
	var cerr [szName]C.char

	if ret == C.EFML_SUCCESS {
		return nil
	}

	C.EfmlErrorString(ret, &cerr[0])
	err := C.GoString(&cerr[0])

	return fmt.Errorf("%v", err)
}

func (h Handle) GetLogicId() (uint, error) {
	var logic_id C.uint
	r := C.EfmlGetDevLogicId(C.uint(h.Dev_Idx), &logic_id)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		fmt.Println("can't find dev logic node:", errorString(r).Error())
		return 0, errorString(r)
	}
	return uint(logic_id), nil
}

func (h Handle) GetBusId() (path string, err error) {

	path = ""
	devInfo, err := h.GetDevInfo()
	if err != nil {
		return
	}
	driverAP, err := GetDriverAccessPoint()

	if err == nil {
		path = fmt.Sprintf("%s%04x:%02x:%02x.%x", driverAP, devInfo.Domain_Id, devInfo.Bus_Id, devInfo.Dev_Id, devInfo.Func_Id)
	}
	return
}

/*
 * @brief Enflame Management Library Initialization.
 * @return efmlReturn_t
 */
func Init() error {
	// SIGUSR1 for PSE efml have special use
	signal.Ignore(syscall.SIGUSR1)
	signal.Ignore(syscall.SIGUSR2)
	return errorString(dl.Init())
}

/*
 * @brief Enflame Management Library Initialization.
 * @return efmlReturn_t
 */
func InitV2(no_driver /* use driver or not */ bool) error {
	// SIGUSR1 for PSE efml have special use
	signal.Ignore(syscall.SIGUSR1)
	signal.Ignore(syscall.SIGUSR2)
	return errorString(dl.InitV2(no_driver))
}

/*
 * @brief Enflame Management Library Shutdown.
 */
func Shutdown() error {
	return errorString(dl.Shutdown())
}

/*
 * @brief Enflame Management Library get driver version info.
 */
func GetDriverVer() (string, error) {
	var ver [szName]C.char

	r := C.EfmlGetDriverVer(&ver[0])
	return C.GoString(&ver[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get driver config and status access path info.
 */
func GetDriverAccessPoint() (string, error) {
	var ver [szName]C.char

	r := C.EfmlGetDriverAccessPoint(&ver[0])
	return C.GoString(&ver[0]), errorString(r)
}

/**
 * @brief Enflame Management Library get the total number of special device's clusters.
 *
 */
func getClusterCount_v1(dev_idx uint) (uint, error) {
	var cluster_cnt C.uint
	r := C.EfmlGetClusterCount(C.uint(dev_idx), &cluster_cnt)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(cluster_cnt), errorString(r)
}

func (h Handle) GetClusterCount() (uint, error) {
	return getClusterCount_v1(h.Dev_Idx)
}

/*
 * @brief Enflame Management Library get the total number of supported devices.
 */
func getDevCount_v1() (uint, error) {
	var dev_cnt C.uint
	r := C.EfmlGetDevCount(&dev_cnt)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(dev_cnt), errorString(r)
}

func GetDevCountFromEfml() (uint, error) {
	dev_cnt, err := getDevCount_v1()
	return uint(dev_cnt), err
}

func GetDevCount() (uint, error) {
	var dev_cnt C.uint
	r := C.EfmlGetDevCount(&dev_cnt)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(dev_cnt), errorString(r)
}

/*
 * @brief Enflame Management Library get the device name.
 */
func (h Handle) GetDevName() (string /* device_Name */, error) {
	var name [szName]C.char

	r := C.EfmlGetDevName(C.uint(h.Dev_Idx), &name[0])
	return C.GoString(&name[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get the slot or OAM name.
 */
func (h Handle) GetDevSlotOamName() (string /* slot_Name */, error) {
	var name [szName]C.char

	r := C.EfmlGetDevSlotOamName(C.uint(h.Dev_Idx), &name[0])
	return C.GoString(&name[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get the device temperature.
 */
func (h Handle) GetDevTemp() (thermalInfo *DevThermalInfo, err error) {
	var thermal C.efmlDevThermalInfo_t

	r := C.EfmlGetDevTemp(C.uint(h.Dev_Idx), &thermal)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	thermalInfo = &DevThermalInfo{
		Cur_Dev_Temp:  float32(thermal.cur_dev_temp),
		Cur_Hbm0_Temp: float32(thermal.cur_hbm0_temp),
		Cur_Hbm1_Temp: float32(thermal.cur_hbm1_temp),
	}
	return
}

/*
 * @brief Enflame Management Library get the device current power consumption.
 */
func (h Handle) GetDevPwr() (powerInfo *DevPowerInfo, err error) {
	var power C.efmlDevPowerInfo_t

	r := C.EfmlGetDevPwr(C.uint(h.Dev_Idx), &power)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	powerInfo = &DevPowerInfo{
		Pwr_Capability:      float32(power.pwr_capability),
		Cur_Pwr_Consumption: float32(power.cur_pwr_consumption),
	}
	return
}

/*
 * @brief Enflame Management Library get the device DPM level.
 */
func (h Handle) GetDevDpmLevel() (uint, error) {
	var dpm_Level C.uint
	r := C.EfmlGetDevDpmLevel(C.uint(h.Dev_Idx), &dpm_Level)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(dpm_Level), errorString(r)
}

/*
 * @brief Enflame Management Library get the device mem info.
 */
func (h Handle) GetDevMem() (memInfo *DevMemInfo, err error) {
	var mem C.efmlDevMemInfo_t

	r := C.EfmlGetDevMem(C.uint(h.Dev_Idx), &mem)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	memInfo = &DevMemInfo{
		Mem_Total_Size: uint(mem.mem_total_size),
		Mem_Used:       uint(mem.mem_used),
	}
	return
}

/*
 * @brief Enflame Management Library get the device usage info.
 */
func (h Handle) GetDevDtuUsage() (float32, error) {
	var usage C.float
	r := C.EfmlGetDevDtuUsage(C.uint(h.Dev_Idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return float32(usage), errorString(r)
}

/*
 * @brief Enflame Management Library get the device usage info by sampling and statistics.
 */
func (h Handle) GetDevDtuUsageAsync() (float32, error) {
	var usage C.float
	r := C.EfmlGetDevDtuUsageAsync(C.uint(h.Dev_Idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return float32(usage), errorString(r)
}

/**
 * @brief Enflame Management Library get the device cluster usage.
 *
 */
func (h Handle) GetClusterUsage(cluster_idx uint) (float32, error) {
	var usage C.float
	r := C.EfmlGetDevClusterUsage(C.uint(h.Dev_Idx), C.uint(cluster_idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return float32(usage), errorString(r)
}

/**
 * @brief Enflame Management Library get the cluster hbm memory.
 *
 */
func (h Handle) GetDevClusterHbmMem(cluster_idx uint) (memInfo *ClusterHbmMemInfo, err error) {
	var mem C.efmlClusterHbmMemInfo_t

	r := C.EfmlGetDevClusterHbmMem(C.uint(h.Dev_Idx), C.uint(cluster_idx), &mem)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	memInfo = &ClusterHbmMemInfo{
		Mem_Total_Size: uint(mem.mem_total_size),
		Mem_Used:       uint(mem.mem_used),
	}
	return
}

/*
 * @brief Enflame Management Library get the device clock info.
 */
func (h Handle) GetDevClk() (clkInfo *DevClkInfo, err error) {
	var clk C.efmlDevClkInfo_t

	r := C.EfmlGetDevClk(C.uint(h.Dev_Idx), &clk)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	clkInfo = &DevClkInfo{
		Cur_Hbm_Clock: uint(clk.cur_hbm_clock),
		Cur_Dtu_Clock: uint(clk.cur_dtu_clock),
	}
	return
}

/*
 * @brief Enflame Management Library get device info.
 */
func (h Handle) GetDevInfo() (devInfo *DeviceInfo, err error) {
	var dev C.efmlDeviceInfo_t

	r := C.EfmlGetDevInfo(C.uint(h.Dev_Idx), &dev)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	devInfo = &DeviceInfo{
		Name:      C.GoString(&dev.name[0]),
		Vendor_Id: uint(dev.vendor_id),
		Device_Id: uint(dev.device_id),
		Domain_Id: uint(dev.domain_id),
		Bus_Id:    uint(dev.bus_id),
		Dev_Id:    uint(dev.dev_id),
		Func_Id:   uint(dev.func_id),
	}
	return
}

/*
 * @brief Enflame Management Library get firmware version info.
 */
func (h Handle) GetFwVersion() (string, error) {
	var ver [szName]C.char

	r := C.EfmlGetFwVersion(C.uint(h.Dev_Idx), &ver[0])
	return C.GoString(&ver[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get device UUID info.
 */
func (h Handle) getDevUuid_v1() (string /* uuid */, error) {
	var uuid [szUUID]C.char

	r := C.EfmlGetDevUuid(C.uint(h.Dev_Idx), &uuid[0])
	return C.GoString(&uuid[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get the device pg count.
 */
func (h Handle) GetDevPGCount() (uint, error) {
	var pg_cnt C.uint
	r := C.EfmlGetPGCount(C.uint(h.Dev_Idx), &pg_cnt)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(pg_cnt), errorString(r)
}

/**
 * @brief Enflame Management Library get the pg usage.
 *
 */
func (h Handle) GetPGUsage(pg_idx uint) (float32, error) {
	var usage C.float
	r := C.EfmlGetDevPGUsage(C.uint(h.Dev_Idx), C.uint(pg_idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return float32(usage), errorString(r)
}

/**
 * @brief Enflame Management Library get the pg usage by sampling and statistics.
 *
 */
func (h Handle) GetPGUsageAsync(pg_idx uint) (float32, error) {
	var usage C.float
	r := C.EfmlGetDevPGUsageAsync(C.uint(h.Dev_Idx), C.uint(pg_idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return float32(usage), errorString(r)
}

/**
 * @brief Enflame Management Library get event message.
 *
 */
func GetEvent(timeout_ms int) (event_info *EventInfo, err error) {
	var event C.efmlEvent_t
	r := C.EfmlGetEvent(C.int(timeout_ms), &event)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	} else if r == C.EFML_ERROR_TIMEOUT {
		err = errorString(r)
		return nil, err
	}

	err = errorString(r)

	event_info = &EventInfo{
		Id:   uint(C.uint(event.event_id)),
		Type: uint(C.uint(event.event_type)),
		Msg:  C.GoString(&event.event_msg[0]),
	}
	return
}

func (h Handle) getDevUuid_v2() (string, error) {
	filePath, _ := h.GetBusId()
	filePath += "/ssm/chipid"
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("can't open file:", err.Error())
		return strconv.Itoa(0), err
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	return string(line), err
}

func (h Handle) GetDevUuidFromEfml() (string, error) {
	return h.getDevUuid_v1()
}

func (h Handle) GetDevUuidFromDriver() (string, error) {
	return h.getDevUuid_v2()
}

func (h Handle) GetDevUuid() (string, error) {
	filePath, _ := h.GetBusId()
	filePath += "/ssm/chipid"

	if _, err := os.Lstat(filePath); err == nil {
		return h.getDevUuid_v2()
	} else {
		return h.getDevUuid_v1()
	}
}

/*
 * @brief Enflame Management Library select one target device by index.
 */
func (h Handle) SelDevByIndex() error {
	return errorString(C.EfmlSelDevByIndex(C.uint(h.Dev_Idx)))
}

/*
 * @brief Enflame Management Library get current device pcie link speed.
 */
func (h Handle) GetPcieLinkSpeed() (uint, error) {
	var linkSpeed C.efmlPcieSpeed_t

	r := C.EfmlGetPcieLinkSpeed(C.uint(h.Dev_Idx), &linkSpeed)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(linkSpeed), errorString(r)
}

/*
 * @brief Enflame Management Library get current device pcie link width.
 */
func (h Handle) GetPcieLinkWidth() (uint, error) {
	var linkWidth C.efmlPcieWidth_t

	r := C.EfmlGetPcieLinkWidth(C.uint(h.Dev_Idx), &linkWidth)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(linkWidth), errorString(r)
}

/*
 * @brief Enflame Management Library get current device pcie link info.
 */
func (h Handle) GetPcieLinkInfo() (linkInfo *LinkInfo, err error) {
	var pcie_Linkinfo C.efmlPcieLinkInfo_t

	r := C.EfmlGetPcieLinkInfo(C.uint(h.Dev_Idx), &pcie_Linkinfo)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	linkInfo = &LinkInfo{
		Link_Speed:     uint(C.uint(pcie_Linkinfo.link_speed)),
		Max_Link_Speed: uint(C.uint(pcie_Linkinfo.max_link_speed)),
		Link_Width:     uint(C.uint(pcie_Linkinfo.link_width)),
		Max_Link_Width: uint(C.uint(pcie_Linkinfo.max_link_width)),
	}
	return
}

/*
 * @brief Enflame Management Library get pcie throughput info.
 */
func (h Handle) GetPcieThroughput() (throughputInfo *ThroughputInfo, err error) {
	var throughPut C.efmlPcieThroughputInfo_t

	r := C.EfmlGetPcieThroughput(C.uint(h.Dev_Idx), &throughPut)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	throughputInfo = &ThroughputInfo{
		Tx_Throughput: float32(throughPut.tx_throughput),
		Rx_Throughput: float32(throughPut.rx_throughput),
		Tx_Nak:        uint64(throughPut.tx_nak),
		Rx_Nak:        uint64(throughPut.rx_nak),
	}
	return
}

/*
 * @brief Enflame Management Library get dtu rma status.
 */
func (h Handle) GetDevRmaStatus() (rmaStatus *DevRmaStatus, err error) {
	var rma C.efmlRmaStatus_t
	r := C.EfmlGetDevRmaStatus(C.uint(h.Dev_Idx), &rma)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	rmaStatus = &DevRmaStatus{
		SupportRma: bool(rma.is_dtu_support_rma),
		Flags:      bool(rma.flags),
	}
	return
}

/*
 * @brief Enflame Management Library get dtu rma details.
 */
func (h Handle) GetDevRmaDetails() (rmaDetails *DevRmaDetails, err error) {
	var rma C.efmlRmaDetails_t
	r := C.EfmlGetDevRmaDetails(C.uint(h.Dev_Idx), &rma)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	rmaDetails = &DevRmaDetails{
		SupportRma: bool(rma.is_dtu_support_rma),
		Flags:      bool(rma.flags),
		Dbe:        uint(rma.dbe_count),
	}
	return
}

/*
 * @brief Enflame Management Library get dram ECC status.
 */
func (h Handle) GetDevEccStatus() (eccStatus *DevEccStatus, err error) {
	var ecc C.efmlEccStatus_t
	r := C.EfmlGetDevEccStatus(C.uint(h.Dev_Idx), &ecc)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	eccStatus = &DevEccStatus{
		Enabled: bool(ecc.enabled),
		Pending: bool(ecc.pending),
		Pdblack: bool(ecc.pdblack),
		Ecnt_sb: uint(ecc.ecnt_sb),
		Ecnt_db: uint(ecc.ecnt_db),
	}
	return
}

/*
 * @brief Enflame Management Library get total ccix port numbers.
 */
func (h Handle) GetEslPortNum() (uint, error) {
	var num C.uint
	r := C.EfmlGetEslPortNum(C.uint(h.Dev_Idx), &num)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(num), errorString(r)
}

/*
 * @brief Enflame Management Library get ccix port info.
 */
func (h Handle) GetEslPortInfo(port_id uint) (portInfo *EslPortInfo, err error) {
	var ccixPort C.efmlEslPortInfo_t

	r := C.EfmlGetEslPortInfo(C.uint(h.Dev_Idx), C.uint(port_id), &ccixPort)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	portInfo = &EslPortInfo{
		Connected: uint(ccixPort.connected),
		Uuid:      C.GoString(&ccixPort.uuid[0]),
		Vendor_Id: uint(ccixPort.vendor_id),
		Device_Id: uint(ccixPort.device_id),
		Domain_Id: uint(ccixPort.domain_id),
		Bus_Id:    uint(ccixPort.bus_id),
		Dev_Id:    uint(ccixPort.dev_id),
		Func_Id:   uint(ccixPort.func_id),
		Port_Id:   uint(ccixPort.port_id),
		Port_Type: uint(C.uint(ccixPort.port_type)),

		Remote_Card_Id:   uint(ccixPort.remote_card_id),
		Remote_Uuid:      C.GoString(&ccixPort.remote_uuid[0]),
		Remote_Vendor_Id: uint(ccixPort.remote_vendor_id),
		Remote_Device_Id: uint(ccixPort.remote_device_id),
		Remote_Domain_Id: uint(ccixPort.remote_domain_id),
		Remote_Bus_Id:    uint(ccixPort.remote_bus_id),
		Remote_Dev_Id:    uint(ccixPort.remote_dev_id),
		Remote_Func_Id:   uint(ccixPort.remote_func_id),
		Remote_Port_Id:   uint(ccixPort.remote_port_id),
		Remote_Port_Type: uint(C.uint(ccixPort.remote_port_type)),
	}
	return
}

/*
 * @brief Enflame Management Library get ccix link info.
 */
func (h Handle) GetEslLinkInfo(port_id uint) (linkInfo *LinkInfo, err error) {
	var ccix_Linkinfo C.efmlEslLinkInfo_t

	r := C.EfmlGetEslLinkInfo(C.uint(h.Dev_Idx), C.uint(port_id), &ccix_Linkinfo)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}
	err = errorString(r)

	linkInfo = &LinkInfo{
		Link_Speed:     uint(C.uint(ccix_Linkinfo.link_speed)),
		Max_Link_Speed: uint(C.uint(ccix_Linkinfo.max_link_speed)),
		Link_Width:     uint(C.uint(ccix_Linkinfo.link_width)),
		Max_Link_Width: uint(C.uint(ccix_Linkinfo.max_link_width)),
	}

	return
}

/*
 * @brief Enflame Management Library get ccix dtuid info.
 */
func (h Handle) GetEslDtuId() (uint, error) {
	var id C.uint
	r := C.EfmlGetEslDtuId(C.uint(h.Dev_Idx), &id)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, nil
	}

	return uint(id), errorString(r)
}

/*
 * @brief Enflame Management Library get ccix throughput info.
 */
func (h Handle) GetEslThroughput(port_id uint) (throughputInfo *ThroughputInfo, err error) {
	var ccixThroughPut C.efmlEslThroughputInfo_t

	r := C.EfmlGetEslThroughput(C.uint(h.Dev_Idx), C.uint(port_id), &ccixThroughPut)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, nil
	}

	err = errorString(r)

	throughputInfo = &ThroughputInfo{
		Tx_Throughput: float32(ccixThroughPut.tx_throughput),
		Rx_Throughput: float32(ccixThroughPut.rx_throughput),
		Tx_Nak:        uint64(ccixThroughPut.tx_nak),
		Rx_Nak:        uint64(ccixThroughPut.rx_nak),
	}
	return
}

func (h Handle) GetSsmFwHeartBeat() (count uint, err error) {
	filePath, _ := h.GetBusId()
	filePath += "/ssm/count"
	if _, err := os.Lstat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("can't open file:", err.Error())
			return 0, err
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		line, _, err := reader.ReadLine()
		ssm_count, _ := strconv.Atoi(string(line))
		return uint(ssm_count), err
	} else {
		return 0, err
	}
}

func (h Handle) GetDevMajorMain() (major uint, main uint, err error) {
	bus_id_path, _ := h.GetBusId()
	logic_id, err := h.GetLogicId()
	if err != nil {
		return 0, 0, err
	}
	filePath := bus_id_path + "/enflame/gcu" + strconv.Itoa(int(logic_id)) + "/dev"
	if _, err := os.Lstat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("can't open file:", err.Error())
			return 0, 0, err
		}
		defer file.Close()
		reader := bufio.NewReader(file)

		line, _, err := reader.ReadLine()
		slice := strings.Split(string(line), ":")
		major, _ := strconv.Atoi(slice[0])
		main, _ := strconv.Atoi(slice[1])
		return uint(major), uint(main), err
	} else {
		return 0, 0, err
	}
}

func (h Handle) GetDevState() (state string, err error) {
	filePath, _ := h.GetBusId()
	filePath += "/device_state"
	if _, err := os.Lstat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("can't open file:", err.Error())
			return "", err
		}
		defer file.Close()
		reader := bufio.NewReader(file)

		line, _, err := reader.ReadLine()
		return string(line), err
	} else {
		return "", err
	}
}

func (h Handle) GetDevInSleepMode() (sleep uint, err error) {
	filePath, _ := h.GetBusId()
	filePath += "/ssm/status"
	if _, err := os.Lstat(filePath); err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("can't open file:", err.Error())
			return 0, err
		}
		defer file.Close()
		r := bufio.NewReader(file)

		line, _, err := r.ReadLine()
		sleep, _ := strconv.Atoi(string(line))
		return uint(sleep), err
	} else {
		return 0, err
	}
}

func GetDeviceHandleByIndex(dev_idx uint) (Handle, error) {
	var h Handle
	h = Handle{
		Dev_Idx: dev_idx,
	}
	return h, nil
}

func (h Handle) GetDevSKU(dev_idx uint) (string, error) {
	var devSKU [szName]C.char

	r := C.EfmlGetDevSKU(C.uint(dev_idx), &devSKU[0])
	return C.GoString(&devSKU[0]), errorString(r)
}

/*
 * @brief Enflame Management Library get the total number of virtual devices per device.
 */
func (h Handle) GetVdevCount() (uint, error) {
	var vdev_cnt C.uint
	r := C.EfmlGetVdevCount(C.uint(h.Dev_Idx), &vdev_cnt)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, errorString(r)
	}

	return uint(vdev_cnt), errorString(r)
}

/*
 * @brief Enflame Management Library get virtual devices index in os.
 */
func (h Handle) GetVdevList() (vdevList []uint, err error) {
	var count C.uint32_t
	var vDevIds [32]C.uint32_t
	r := C.EfmlGetVdevList(C.uint(h.Dev_Idx), &vDevIds[0], &count)
	err = errorString(r)
	for i := uint(0); i < uint(count); i++ {
		vdevList = append(vdevList, uint(vDevIds[i]))
	}

	return
}

/*
 * @brief Enflame Management Library get process info on device.
 */
func (h Handle) GetProcessInfo() (pInfos []ProcessInfo, err error) {
	var count C.uint32_t
	var processInfos [32]C.efmlProcessInfo_t
	r := C.EfmlGetProcessInfo(C.uint(h.Dev_Idx), &count, &processInfos[0])
	err = errorString(r)
	for i := uint(0); i < uint(count); i++ {
		pInfos = append(pInfos, ProcessInfo{
			Pid:         uint(processInfos[i].pid),
			DevMemUsage: uint64(processInfos[i].dev_mem_usage),
			SysMemUsage: uint64(processInfos[i].sys_mem_usage),
		})
	}

	return
}

/*
 * @brief Enflame Management Library get the virtual device mem info.
 */
func (h Handle) GetVdevDtuMem(vdev_idx uint) (memInfo *DevMemInfo, err error) {
	var mem C.efmlDevMemInfo_t

	r := C.EfmlGetVdevMem(C.uint(h.Dev_Idx), C.uint(vdev_idx), &mem)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return nil, errorString(r)
	}

	err = errorString(r)

	memInfo = &DevMemInfo{
		Mem_Total_Size: uint(mem.mem_total_size),
		Mem_Used:       uint(mem.mem_used),
	}
	return
}

/*
 * @brief Enflame Management Library get the virtual device usage.
 */
func (h Handle) GetVdevDtuUsage(vdev_idx uint) (float32, error) {
	var usage C.float
	r := C.EfmlGetVdevDtuUsage(C.uint(h.Dev_Idx), C.uint(vdev_idx), &usage)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return 0, errorString(r)
	}

	return float32(usage), errorString(r)
}
