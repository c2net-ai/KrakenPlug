// Copyright (c) 2022 Enflame. All Rights Reserved.

package eflib

import (
	"fmt"
	"log"
	"os"
	"time"

	"go-eflib/efml"
)

func GetDeviceCount() (uint32, error) {
	count, err := efml.GetDevCount()
	if err != nil {
		log.Printf("Failed to get dev count\n")
		return 0, err
	}
	return uint32(count), nil
}

func GetVDeviceCount() (uint32, error) {
	vcount := 0
	for i := 0; i < GCU_VDEV_MAX; i++ {
		vdevicePath := fmt.Sprintf("%s%d", GCU_VDEV_NAME, i)
		if fileIsExist(vdevicePath) {
			vcount++
		}
	}
	return uint32(vcount), nil
}

func fileIsExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}

/*
 * check devCount use different way and try serial times if call API failed.
 * return nil if result is same
 * return err if result is difference
 */
func CheckDevCountState() (uint, error) {
	var maxTry uint = 2
	var i uint
	for i = 0; i < maxTry; i++ {
		count, err1 := efml.GetDevCountFromEfml()
		pcieDevCount := uint32(count)
		driverDevCount, err2 := GetDeviceCount()
		if (err1 == nil) && (err2 == nil) {
			if pcieDevCount == driverDevCount {
				return i, nil
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
		if err1 != nil {
			fmt.Println("Failed to get device count from pcie info!")
		}
		if err2 != nil {
			fmt.Println("Failed to get device count from the driver!")
		}
	}
	return maxTry, fmt.Errorf("retry max times: %d", maxTry)
}

func GetDeviceClock(h efml.Handle) (uint32, error) {
	clock, err := h.GetDevClk()
	if err != nil {
		log.Printf("get device clock is unsupported\n")
		//return 0, nil
	}
	return uint32(clock.Cur_Dtu_Clock), nil
}

func GetDevicePowerMode(h efml.Handle) (string, error) {
	isLowPower, err := h.GetDevIsLowPowerMode()
	if err != nil {
		log.Printf("Failed to get device low power mode.\n")
		return "", err
	}
	if isLowPower {
		return "Sleep", err
	} else {
		return "Active", err
	}
}

func GetHBMClock(h efml.Handle) (uint32, error) {
	clock, err := h.GetDevClk()
	if err != nil {
		log.Printf("Failed to get hbm clock\n")
		return 0, err
	}
	return uint32(clock.Cur_Hbm_Clock), err
}

func GetDeviceBusID(h efml.Handle) (string, error) {
	busid, err := h.GetBusId()
	if err != nil {
		log.Printf("Failed to get bus id\n")
		return "", err
	}
	return busid, nil
}

func GetDeviceHandle(index uint32) (efml.Handle, error) {
	return efml.GetDeviceHandleByIndex(uint(index))
}

func GetDeviceType(h efml.Handle) (string, error) {
	//h.GetDevInfo()
	devSKU, err := h.GetDevSKU()
	if err != nil {
		log.Printf("Failed to get device Info \n")
		return "", err
	}
	return "Enflame " + devSKU, nil
}

func GetDeviceUUID(h efml.Handle) (string, error) {
	uuid, err := h.GetDevUuid()
	if err != nil {
		log.Printf("Failed to get device uuid\n")
		return "", err
	}
	gcuuuid := "GCU-" + uuid

	return gcuuuid, err
}

func GetDeviceMinor(h efml.Handle) (uint32, error) {
	_, minor, err := h.GetDevMajorMain()
	if err != nil {
		log.Printf("Failed to get device minor\n")
		return 0, err
	}
	return uint32(minor), err
}

func GetDeviceName(h efml.Handle) (string, error) {
	_, minor, err := h.GetDevMajorMain()
	if err != nil {
		log.Printf("Failed to get device name\n")
		return "", err
	}
	name := fmt.Sprintf("%s%d", "gcu", minor)
	return name, nil
}

func GetDeviceSKU(h efml.Handle) (string, error) {
	devSKU, err := h.GetDevSKU()
	if err != nil {
		log.Printf("Failed to get device SKU\n")
		return "", err
	}
	return devSKU, nil
}

func GetDeviceSlotNumber(h efml.Handle) (string, error) {
	slot, err := h.GetDevSlotOamName()
	if err != nil {
		log.Printf("Failed to get device slot number\n")
		return "", err
	}
	return slot, nil
}

func HasEslLink(h efml.Handle) (bool, error) {
	// Will be replaced by h.HasEslLink() provided by efml.
	devSKU, err := GetDeviceSKU(h)
	if err != nil {
		return false, err
	}

	if devSKU == "T10" || devSKU == "T10s" || devSKU == "T11" ||
		devSKU == "T20" || devSKU == "T21" {
		return true, nil
	}
	return false, nil
}

func GetDeviceTemperature(h efml.Handle) (float32, error) {
	thermalInfo, err := h.GetDevTempV2()
	if err != nil {
		log.Printf("Failed to get device temperature\n")
		return 0, err
	}
	return thermalInfo.Cur_Asic_Temp, nil
}

func GetDevicePowerInfo(h efml.Handle) (float32, float32, float32, error) {
	powerInfo, err := h.GetDevPwr()
	if err != nil {
		log.Printf("Failed to get device power info\n")
		return 0, 0, 0, err
	}
	if powerInfo.Pwr_Capability == 0 {
		log.Printf("The device power capability = 0\n")
		return 0, 0, 0, nil
	}

	powerUsage := powerInfo.Cur_Pwr_Consumption / powerInfo.Pwr_Capability
	powerConsumption := powerInfo.Cur_Pwr_Consumption
	powerCapability := powerInfo.Pwr_Capability

	return powerUsage * 100, powerConsumption, powerCapability, nil
}

func GetDeviceMemoryTotalSizeBytes(h efml.Handle) (uint64, error) {
	memInfo, err := h.GetDevMem()
	if err != nil {
		log.Printf("Failed to get device memory info\n")
		return 0, err
	}

	// 1073741824 = 1024 * 1024 * 1024
	//GBSize := memInfo.Mem_Total_Size / 1073741824

	//return uint32(GBSize), nil
	return uint64(memInfo.Mem_Total_Size), nil
}

func GetDeviceMemoryUsedSizeBytes(h efml.Handle) (uint64, error) {
	memInfo, err := h.GetDevMem()
	if err != nil {
		log.Printf("Failed to get device memory info\n")
		return 0, err
	}

	// 1073741824 = 1024 * 1024 * 1024
	//GBSize := memInfo.Mem_Total_Size / 1073741824
	memorySize := memInfo.Mem_Used * 1024 * 1024

	return uint64(memorySize), nil
}

func GetDeviceMemoryUsage(h efml.Handle) (float32, error) {
	memInfo, err := h.GetDevMem()
	fmt.Println("getDeviceMemoryUsage Used is ", memInfo.Mem_Used, "total is ", memInfo.Mem_Total_Size, "idx is ", h.Dev_Idx)
	if err != nil {
		log.Printf("Failed to get device memory usage info\n")
		return 0, err
	}
	if memInfo.Mem_Total_Size != 0 {
		//Currently, all of memory is allocated by SDK while init
		//memUsage := float32(memInfo.Mem_Total_Size) / float32(memInfo.Mem_Total_Size)
		//fmt.Println("getDeviceMemoryUsage Used is ", memInfo.Mem_Used, "total is ", memInfo.Mem_Total_Size)
		//1048576=1024*1024
		totalsize := memInfo.Mem_Total_Size / 1048576

		memUsage := float32(memInfo.Mem_Used) / float32(totalsize)
		return memUsage, nil
	} else {
		return 0, nil
	}
}

func GetDeviceMemoryTotalGBSize(h efml.Handle) (uint64, error) {
	memInfo, err := h.GetDevMem()
	if err != nil {
		log.Printf("Failed to get device memory info\n")
		return 0, err
	}

	// 1073741824 = 1024 * 1024 * 1024
	GBSize := memInfo.Mem_Total_Size / 1073741824

	return uint64(GBSize), nil
}

func GetDeviceMemoryUsedMBSize(h efml.Handle) (uint64, error) {
	memInfo, err := h.GetDevMem()
	if err != nil {
		log.Printf("Failed to get device memory info\n")
		return 0, err
	}

	//MB
	memorySize := memInfo.Mem_Used

	return uint64(memorySize), nil
}

func GetDeviceMemoryInfo(h efml.Handle) (float32, uint64, uint64, error) {
	memInfo, err := h.GetDevMem()
	if err != nil {
		log.Printf("Failed to get device memory usage info\n")
		return 0, 0, 0, err
	}

	// 1073741824 = 1024 * 1024 * 1024
	//GBSize := memInfo.Mem_Total_Size / 1073741824
	memoryUsedSizeBytes := memInfo.Mem_Used * 1024 * 1024

	// 1073741824 = 1024 * 1024 * 1024
	//GBSize := memInfo.Mem_Total_Size / 1073741824
	//return uint32(GBSize), nil
	memoryTotalSizeBytes := memInfo.Mem_Total_Size

	var memUsage float32
	if memInfo.Mem_Total_Size != 0 {
		//Currently, all of memory is allocated by SDK while init
		//memUsage := float32(memInfo.Mem_Total_Size) / float32(memInfo.Mem_Total_Size)
		//fmt.Println("getDeviceMemoryUsage Used is ", memInfo.Mem_Used, "total is ", memInfo.Mem_Total_Size)
		//1048576=1024*1024
		totalsize := memInfo.Mem_Total_Size / 1048576
		memUsage = float32(memInfo.Mem_Used) / float32(totalsize)
	} else {
		memUsage = 0
	}

	return float32(memUsage), uint64(memoryTotalSizeBytes), uint64(memoryUsedSizeBytes), nil
}

func GetDeviceGcuUsage(h efml.Handle) (float32, error) {
	gcuUsage, err := h.GetDevDtuUsageAsync()
	if err != nil {
		log.Printf("Failed to get device gcu usage\n")
		return 0, err
	}
	return gcuUsage, nil
}

func GetDeviceEslThroughput(h efml.Handle) (eslThroughput []*efml.ThroughputInfo, err error) {
	portNum, err := h.GetEslPortNum()
	if err != nil {
		log.Printf("Failed to get device ESL port number\n")
		return nil, err
	}
	for idx := uint(0); idx < portNum; idx++ {
		throughput, err := h.GetEslThroughput(idx)
		if err != nil {
			log.Printf("Failed to get device ESL throughput info\n")
			return nil, err
		} else {
			eslThroughput = append(eslThroughput, throughput)
		}
	}
	// MB/S
	return eslThroughput, nil
}

func GetDeviceEslLinkInfo(h efml.Handle) (linkInfo []*efml.LinkInfo, err error) {
	portNum, err := h.GetEslPortNum()
	if err != nil {
		log.Printf("Failed to get device ESL port number\n")
		return nil, err
	}
	for idx := uint(0); idx < portNum; idx++ {
		info, err := h.GetEslLinkInfo(idx)
		if err != nil {
			log.Printf("Failed to get device ESL Link info\n")
			return nil, err
		} else {
			linkInfo = append(linkInfo, info)
		}
	}

	return linkInfo, nil
}

func GetDeviceEslInfo(h efml.Handle) (eslThroughput []*efml.ThroughputInfo, linkInfo []*efml.LinkInfo, err error) {
	portNum, err := h.GetEslPortNum()
	if err != nil {
		log.Printf("Failed to get device ESL port number\n")
		return nil, nil, err
	}

	for idx := uint(0); idx < portNum; idx++ {
		throughput, err := h.GetEslThroughput(idx)
		if err != nil {
			log.Printf("Failed to get device ESL throughput info\n")
			return nil, nil, err
		} else {
			eslThroughput = append(eslThroughput, throughput)
		}
	}

	for idx := uint(0); idx < portNum; idx++ {
		info, err := h.GetEslLinkInfo(idx)
		if err != nil {
			log.Printf("Failed to get device ESL Link info\n")
			return nil, nil, err
		} else {
			linkInfo = append(linkInfo, info)
		}
	}

	return eslThroughput, linkInfo, nil
}

func GetDeviceClusterUsage(h efml.Handle) (clusterUsage []float64, err error) {
	clusterNum, err := h.GetClusterCount()
	if err != nil {
		log.Printf("get cluster info is unsupported\n")
		return nil, nil
	}
	for idx := uint(0); idx < clusterNum; idx++ {
		usage, err := h.GetClusterUsage(idx)
		if err != nil {
			log.Printf("Failed to get device cluster usage\n")
			return nil, err
		} else {
			usage64 := float64(usage)
			clusterUsage = append(clusterUsage, usage64)
		}
	}
	return clusterUsage, nil
}

func GetDevicePGUsage(h efml.Handle) (PGUsage []float64, err error) {
	PGNum, err := h.GetDevPGCount()
	if err != nil {
		log.Printf("Failed to get device PG number\n")
		return nil, err
	}
	for idx := uint(0); idx < PGNum; idx++ {
		usage, err := h.GetPGUsageAsync(idx)
		if err != nil {
			log.Printf("Failed to get device PG usage\n")
			return nil, err
		} else {
			usage64 := float64(usage)
			PGUsage = append(PGUsage, usage64)
		}
	}
	return PGUsage, nil
}

func GetDeviceVUsage(h efml.Handle, vidx_list []uint) (gcuUsage []float64, err error) {
	for _, vidx := range vidx_list {
		usage, err := h.GetVdevDtuUsage(vidx)
		if err != nil {
			log.Printf("Failed to get device cluster usage\n")
			return nil, err
		} else {
			usage64 := float64(usage)
			gcuUsage = append(gcuUsage, usage64)
		}
	}
	return gcuUsage, nil
}

func GetVIndexList(h efml.Handle) (vindex_list []uint, err error) {
	vindex_list, err = h.GetVdevList()
	if err != nil {
		log.Printf("Failed to get device vgcu indexs\n")
		return nil, err
	}
	return vindex_list, nil
}

func GetDeviceVMem(h efml.Handle, vindex_list []uint) (gcuVMemUsed []float64, gcuVMemSize []float64, gcuVMemUsage []float64, err error) {
	for _, vidx := range vindex_list {
		meminfo, err := h.GetVdevDtuMem(vidx)
		if err != nil {
			log.Printf("Failed to get device vgcu mem\n")
			return nil, nil, nil, err
		} else {
			vmemUsed := meminfo.Mem_Used * 1024 * 1024
			gcuVMemUsed = append(gcuVMemUsed, float64(vmemUsed))
			vmemSize := meminfo.Mem_Total_Size
			gcuVMemSize = append(gcuVMemSize, float64(vmemSize))
			vmemUsage := float64(0)
			if meminfo.Mem_Total_Size != 0 {
				vmemUsage = float64(meminfo.Mem_Used) * 1024 * 1024 / float64(meminfo.Mem_Total_Size)
			}
			gcuVMemUsage = append(gcuVMemUsage, float64(vmemUsage))
		}
	}
	return gcuVMemUsed, gcuVMemSize, gcuVMemUsage, nil
}

/*
	type DevRmaStatus struct {
	        SupportRma bool
	        Flags      bool
	}
*/
func GetDeviceRmaStatus(h efml.Handle) (*efml.DevRmaStatus, error) {
	rmaStatus, err := h.GetDevRmaStatus()
	if err != nil {
		log.Printf("Failed to get device rma status\n")
		return nil, err
	}
	return rmaStatus, nil
}

/*
	type DevEccStatus struct {
	        Enabled bool
	        Pending bool
	        Pdblack bool
	        Ecnt_sb uint
	        Ecnt_db uint
	}
*/
func GetDeviceEccStatus(h efml.Handle) (*efml.DevEccStatus, error) {
	eccStatus, err := h.GetDevEccStatus()
	if err != nil {
		log.Printf("Failed to get device ecc status\n")
		return nil, err
	}
	return eccStatus, nil
}

func GetDevicePcieLinkSpeed(h efml.Handle) (uint, error) {
	linkSpeed, err := h.GetPcieLinkSpeed()
	if err != nil {
		log.Printf("Failed to get device pcie link speed\n")
		return 0, err
	}
	return linkSpeed, nil
}

func GetDevicePcieLinkWidth(h efml.Handle) (uint, error) {
	linkWidth, err := h.GetPcieLinkWidth()
	if err != nil {
		log.Printf("Failed to get device pcie link width\n")
		return 0, err
	}
	return linkWidth, nil
}

func GetDevicePcieLinkInfo(h efml.Handle) (*efml.LinkInfo, error) {
	linkInfo, err := h.GetPcieLinkInfo()
	if err != nil {
		log.Printf("Failed to get device pcie info\n")
		return nil, err
	}
	return linkInfo, nil
}

func GetEventInfo(timeout_ms int) (*efml.EventInfo, error) {
	eventInfo, err := efml.GetEvent(timeout_ms)
	if err != nil {
		log.Printf("Failed to get event info\n")
		return nil, err
	}

	return eventInfo, nil
}
