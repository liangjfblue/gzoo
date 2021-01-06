package monitor

import (
	"log"
	"os"
	"runtime"
	"strconv"

	sigar "github.com/cloudfoundry/gosigar"
)

type MachineInfo struct {
	pid     int
	mem     sigar.Mem
	pMem    runtime.MemStats
	cpuSize int
	cpu     sigar.Cpu
	pCpu    sigar.ProcCpu
}

func newMachine() *MachineInfo {
	m := &MachineInfo{
		pid:     os.Getpid(),
		cpuSize: runtime.NumCPU(),
	}
	_ = m.cpu.Get()
	return m
}

func (info *MachineInfo) GetInfo() (r map[string]interface{}) {
	var err error
	r = make(map[string]interface{})

	err = info.mem.Get()
	runtime.ReadMemStats(&info.pMem)
	cpuOld := info.cpu
	err = info.cpu.Get()
	cpuDiff := info.cpu.Delta(cpuOld)
	err = info.pCpu.Get(info.pid)
	r["systemCpuSize"] = strconv.FormatInt(int64(info.cpuSize), 10)

	sysCpuPct := float64(cpuDiff.Total()-cpuDiff.Idle) / float64(cpuDiff.Total())
	r["systemCpuUsage"] = strconv.FormatFloat(sysCpuPct, 'f', 4, 64)
	r["systemMemSize"] = strconv.FormatUint(info.mem.Total, 10)
	r["systemMemUsage"] = strconv.FormatUint(info.mem.Used, 10)

	procCpuPct := info.pCpu.Percent / float64(info.cpuSize)
	r["processCpuUsage"] = strconv.FormatFloat(procCpuPct, 'f', 4, 64)
	r["processMemSize"] = strconv.FormatUint(info.pMem.Sys, 10)
	r["processMemUsage"] = strconv.FormatUint(info.pMem.Alloc, 10)

	if err != nil {
		log.Println(err)
	}
	return
}
