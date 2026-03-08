package apis

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	machinecode "github.com/super-l/machine-code"
)

func init() {

	api.Register("GET", "dash/cpu-info", cpuInfo)
	api.Register("GET", "dash/cpu", cpuStats)
	api.Register("GET", "dash/memory", memStats)
	api.Register("GET", "dash/disk", diskStats)
	api.Register("GET", "dash/net", netStats)
	api.Register("GET", "dash/machine", machineInfo)
}

type MemStat struct {
	Total     uint64  `json:"total"`
	Available uint64  `json:"available"`
	Used      uint64  `json:"used"`
	Percent   float64 `json:"percent"`
	Free      uint64  `json:"free"`
}

func memStats(ctx *gin.Context) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		api.Error(ctx, err)
		return
	}
	st := &MemStat{
		Total:     stat.Total,
		Available: stat.Available,
		Used:      stat.Used,
		Percent:   stat.UsedPercent,
		Free:      stat.Free,
	}
	api.OK(ctx, st)
}

func cpuInfo(ctx *gin.Context) {
	info, err := cpu.Info()
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if len(info) == 0 {
		api.Fail(ctx, "查询失败")
		return
	}
	api.OK(ctx, info[0])
}

func cpuStats(ctx *gin.Context) {
	//times, err := cpu.Times(false)
	times, err := cpu.Percent(time.Millisecond*200, false)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if len(times) == 0 {
		api.Fail(ctx, "查询失败")
		return
	}
	api.OK(ctx, int(times[0]))
}

type DiskStat struct {
	Path    string  `json:"path"`
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

func diskStats(ctx *gin.Context) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	var usages []*DiskStat
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			api.Error(ctx, err)
			return
		}
		usages = append(usages, &DiskStat{
			Path:    usage.Path,
			Total:   usage.Total,
			Free:    usage.Free,
			Used:    usage.Used,
			Percent: usage.UsedPercent,
		})
	}
	api.OK(ctx, usages)
}

type NetStat struct {
	Index           int      `json:"index"`
	MTU             int      `json:"mtu"`
	Name            string   `json:"name"`
	HardwareAddress string   `json:"hardware_address"`
	Address         []string `json:"address"`
	BytesSent       uint64   `json:"bytes_sent"`
	BytesRecv       uint64   `json:"bytes_recv"`
	PacketsSent     uint64   `json:"packets_sent"`
	PacketsRecv     uint64   `json:"packets_recv"`
}

func netStats(ctx *gin.Context) {
	ns, err := net.Interfaces()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	usages, err := net.IOCounters(true)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	var ss []*NetStat
	for i, n := range ns {
		st := &NetStat{
			Index:           n.Index,
			MTU:             n.MTU,
			Name:            n.Name,
			HardwareAddress: n.HardwareAddr,
		}
		ss = append(ss, st)

		for _, a := range n.Addrs {
			st.Address = append(st.Address, a.Addr)
		}

		if i < len(usages) {
			u := usages[i]
			//补充统计信息
			st.BytesSent = u.BytesSent
			st.BytesRecv = u.BytesRecv
			st.PacketsSent = u.PacketsSent
			st.PacketsRecv = u.PacketsRecv
		}
	}

	api.OK(ctx, ss)
}

func machineInfo(ctx *gin.Context) {
	if machinecode.MachineErr != nil {
		//fmt.Println("获取机器码信息错误:" + machinecode.MachineErr.Error())
		api.Fail(ctx, "error")
		return
	}
	api.OK(ctx, &machinecode.Machine)
}
