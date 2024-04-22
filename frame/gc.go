package frame

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
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

	UpdateMemoryUsageLimitBytes(ctrl.MemoryUsageLimitBytes)
	if ctrl.MemoryUsageLimitPercentage != "" {
		if err := UpdateMemoryUsageLimitPercentage(ctrl.MemoryUsageLimitPercentage, 0); err != nil {
			getLoggerInst().WarningF("Failed to update the percentage of memory usage limit, %v", err)
		}
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

func UpdateMemoryUsageLimitPercentage(usageLimitPercentage string, maxUsageLimitSize uint64) (retErr error) {
	spList := strings.Split(usageLimitPercentage, "%")
	if len(spList) != 2 {
		retErr = fmt.Errorf("format error")
		return
	}

	rate, convErr := strconv.ParseFloat(spList[0], 64)
	if convErr != nil {
		retErr = fmt.Errorf("convert error")
		return
	}
	limitRate := float64(rate) / 100

	if maxUsageLimitSize <= 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		memInfo, errGetMemInfo := mem.VirtualMemoryWithContext(ctx)
		cancel()
		if errGetMemInfo != nil {
			retErr = errGetMemInfo
			return
		}
		maxUsageLimitSize = memInfo.Total
	}

	nowLimitSize := int64(float64(maxUsageLimitSize) * limitRate)
	beforeLimitSize := debug.SetMemoryLimit(nowLimitSize)
	getLoggerInst().InfoF("The usage limit for memory size has been updated from (%.2fGBs/%dMBs/%dBKs/%dBytes) to (%.2fGBs/%dMBs/%dKBs/%dBytes), "+
		"MaxUsageLimitSize: %.2fGBs/%dMBs/%dKBs/%dBytes",
		float64(beforeLimitSize)/float64(1024*1024*1024), beforeLimitSize/1024/1024, beforeLimitSize/1024, beforeLimitSize,
		float64(nowLimitSize)/float64(1024*1024*1024), nowLimitSize/1024/1024, nowLimitSize/1024, nowLimitSize,
		float64(maxUsageLimitSize)/float64(1024*1024*1024), maxUsageLimitSize/1024/1024, maxUsageLimitSize/1024, maxUsageLimitSize)
	return
}
