package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo"
)

// ApplicationInfo contains info about the app
type ApplicationInfo struct {
	Version   string
	GOVersion string
	BuildDate string
	StartTime time.Time
}

// MemoryInfo some memory stats
type MemoryInfo struct {
	// Bytes of allocated heap objects.
	Alloc uint64 `json:"allocated"`

	// Cumulative bytes allocated for heap objects.
	TotalAlloc uint64 `json:"totalAllocated"`

	// The total bytes of memory obtained from the OS.
	System uint64 `json:"system"`

	// The number of pointer lookups performed by the runtime.
	Lookups uint64 `json:"lookups"`

	// The cumulative count of heap objects allocated.
	// The number of live objects is Mallocs - Frees.
	Mallocs uint64 `json:"mallocs"`

	// The cumulative count of heap objects freed.
	Frees uint64 `json:"frees"`

	// The number of completed GC cycles.
	NumGC uint32 `json:"numGC"`
}

// InfoResponse is the info response
type InfoResponse struct {
	// Version of the application
	Version string `json:"version"`

	// Go runtime version
	RuntimeVersion string `json:"runtimeVersion"`

	// The build date of the server application
	BuildDate string `json:"buildDate"`

	// Uptime of the server
	Uptime string `json:"uptime"`

	// Memory info
	MemoryInfo *MemoryInfo `json:"memoryInfo,omitempty"`
}

// The InfoAPI provides handlers for managing projects.
type InfoAPI struct {
	Info ApplicationInfo
}

// GetApplicationInfo gets application info
func (api *InfoAPI) GetApplicationInfo(ctx echo.Context) error {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	ir := InfoResponse{
		Version:        api.Info.Version,
		RuntimeVersion: api.Info.GOVersion,
		BuildDate:      api.Info.BuildDate,
		Uptime:         time.Since(api.Info.StartTime).String(),
		MemoryInfo: &MemoryInfo{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			System:     memStats.Sys,
			Lookups:    memStats.Lookups,
			Mallocs:    memStats.Mallocs,
			Frees:      memStats.Frees,
			NumGC:      memStats.NumGC,
		},
	}

	return ctx.JSON(http.StatusOK, ir)
}
