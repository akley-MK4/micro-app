package frame

import (
	"runtime"
)

var (
	initialMemorySnapshot = &MemorySnapshot{}
)

func setInitialMemorySnapshot() {
	initialMemorySnapshot.update()
}

func GetInitialMemorySnapshot() MemorySnapshot {
	return *initialMemorySnapshot
}

type MemorySnapshot struct {
	AllocBytes    uint64
	AllocMBs      uint64
	TotalAllocMBs uint64
	SysMBs        uint64
	MallocsMbs    uint64
	HeapAllocMBs  uint64
	HeapSysMbs    uint64
}

func (t *MemorySnapshot) update() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	mbSize := uint64(1024 * 1024)
	t.AllocBytes = ms.Alloc
	t.AllocMBs = ms.Alloc / mbSize
	t.TotalAllocMBs = ms.TotalAlloc / mbSize
	t.SysMBs = ms.Sys / mbSize
	t.MallocsMbs = ms.Mallocs / mbSize
	t.HeapAllocMBs = ms.HeapAlloc / mbSize
	t.HeapSysMbs = ms.HeapSys / mbSize
}
