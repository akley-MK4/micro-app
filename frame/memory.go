package frame

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	KBSize = 1024
	MBSize = KBSize * 1024
	GBSize = MBSize * 1024
)

type ForcePolicy struct {
	IntervalMS int `json:"interval_ms"`
	MemPeak    int `json:"mem_peak"`
}

func setGCPolicy(ctrl GCControl) {
	if ctrl.Percent > 0 {
		debug.SetGCPercent(ctrl.Percent)
		getLoggerInst().InfoF("Set GCPercentage to %d", ctrl.Percent)
	}

	if ctrl.DisableDefaultGC {
		debug.SetGCPercent(-1)
		getLoggerInst().Warning("Default GC turned off")
	}

	UpdateMemoryUsageLimitBytes(ctrl.MemoryUsageLimitBytes)

	if !ctrl.EnableForce {
		return
	}
	// Manually regulating GC
	go func() {
		for {
			time.Sleep(time.Duration(ctrl.ForcePolicy.IntervalSecondS) * time.Second)
			runtime.GC()
		}
	}()

	// todo: check mem peak

	getLoggerInst().InfoF("Forced GC enabled, IntervalSecondS: %d, MemPeak: %d",
		ctrl.ForcePolicy.IntervalSecondS, ctrl.ForcePolicy.MemPeak)
}

func UpdateMemoryUsageLimitBytes(limitBytes int64) {
	if limitBytes <= 0 {
		return
	}

	beforeLimitBytes := debug.SetMemoryLimit(limitBytes)
	getLoggerInst().InfoF("The usage limit for memory size has been updated from %d to %d",
		beforeLimitBytes, limitBytes)
	return
}

func PrintCurrentMemorySnapshot() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	type Info struct {
		Alloc      string
		TotalAlloc string
		Sys        string
		Mallocs    string
		HeapAlloc  string
		HeapSys    string
		StackSys   string
	}

	info := Info{
		Alloc:      MemorySizeToString(memStats.Alloc),
		TotalAlloc: MemorySizeToString(memStats.TotalAlloc),
		Sys:        MemorySizeToString(memStats.Sys),
		Mallocs:    MemorySizeToString(memStats.Mallocs),
		HeapAlloc:  MemorySizeToString(memStats.HeapAlloc),
		HeapSys:    MemorySizeToString(memStats.HeapSys),
		StackSys:   MemorySizeToString(memStats.StackSys),
	}

	d, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err.Error()
	}

	return string(d)
}

func MemorySizeToString(size uint64) string {
	return fmt.Sprintf("(%.2fGBs/%.2fMBs/%.2fKBs/%dBytes)",
		float64(size)/float64(GBSize), float64(size)/float64(MBSize), float64(size)/float64(KBSize), size)
}
