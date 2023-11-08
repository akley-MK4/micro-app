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
	if ctrl.DisableDefaultGC {
		debug.SetGCPercent(-1)
		getLoggerInst().Warning("Default GC turned off")
	}

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

	getLoggerInst().WarningF("Forced GC enabled, IntervalSecondS: %d, MemPeak: %d",
		ctrl.ForcePolicy.IntervalSecondS, ctrl.ForcePolicy.MemPeak)
}
