package frame

import (
	"runtime"
	"runtime/debug"
	"time"
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

	UpdateMemorySizeUsageLimit(ctrl.MemorySizeUsageLimit)

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

func UpdateMemorySizeUsageLimit(memorySizeLimit int64) bool {
	if memorySizeLimit < 0 {
		return false
	}

	beforeMemorySizeLimit := debug.SetMemoryLimit(memorySizeLimit)
	getLoggerInst().InfoF("The usage limit for memory size has been updated from %d to %d",
		beforeMemorySizeLimit, memorySizeLimit)
	return true
}
