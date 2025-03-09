package profiling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
	_ "net/http/pprof"
)

type Profiler struct {
	ctx           context.Context
	cancel        context.CancelFunc
	enabled       bool
	profileDir    string
	
	cpuProfile    *os.File
	memProfile    *os.File
	traceFile     *os.File
	blockProfile  *os.File
	mutexProfile  *os.File
	
	startTime     time.Time
	mu            sync.Mutex
}

type ProfileConfig struct {
	EnableCPU     bool
	EnableMemory  bool
	EnableTrace   bool
	EnableBlock   bool
	EnableMutex   bool
	ProfileDir    string
	Duration      time.Duration
}

func NewProfiler(ctx context.Context, config ProfileConfig) *Profiler {
	ctx, cancel := context.WithCancel(ctx)
	
	profiler := &Profiler{
		ctx:        ctx,
		cancel:     cancel,
		enabled:    config.EnableCPU || config.EnableMemory || config.EnableTrace,
		profileDir: config.ProfileDir,
		startTime:  time.Now(),
	}
	
	if profiler.profileDir == "" {
		profiler.profileDir = "./profiles"
	}
	
	os.MkdirAll(profiler.profileDir, 0755)
	
	return profiler
}

func (p *Profiler) Start() error {
	if !p.enabled {
		return nil
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.cpuProfile == nil {
		cpuFile, err := os.Create(fmt.Sprintf("%s/cpu.prof", p.profileDir))
		if err != nil {
			return fmt.Errorf("failed to create CPU profile: %w", err)
		}
		p.cpuProfile = cpuFile
		pprof.StartCPUProfile(cpuFile)
	}
	
	if p.memProfile == nil {
		memFile, err := os.Create(fmt.Sprintf("%s/memory.prof", p.profileDir))
		if err != nil {
			return fmt.Errorf("failed to create memory profile: %w", err)
		}
		p.memProfile = memFile
	}
	
	if p.traceFile == nil {
		traceFile, err := os.Create(fmt.Sprintf("%s/trace.out", p.profileDir))
		if err != nil {
			return fmt.Errorf("failed to create trace file: %w", err)
		}
		p.traceFile = traceFile
		trace.Start(traceFile)
	}
	
	if p.blockProfile == nil {
		blockFile, err := os.Create(fmt.Sprintf("%s/block.prof", p.profileDir))
		if err != nil {
			return fmt.Errorf("failed to create block profile: %w", err)
		}
		p.blockProfile = blockFile
		runtime.SetBlockProfileRate(1)
	}
	
	if p.mutexProfile == nil {
		mutexFile, err := os.Create(fmt.Sprintf("%s/mutex.prof", p.profileDir))
		if err != nil {
			return fmt.Errorf("failed to create mutex profile: %w", err)
		}
		p.mutexProfile = mutexFile
		runtime.SetMutexProfileFraction(1)
	}
	
	return nil
}

func (p *Profiler) Stop() error {
	if !p.enabled {
		return nil
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.cpuProfile != nil {
		pprof.StopCPUProfile()
		p.cpuProfile.Close()
		p.cpuProfile = nil
	}
	
	if p.traceFile != nil {
		trace.Stop()
		p.traceFile.Close()
		p.traceFile = nil
	}
	
	if p.memProfile != nil {
		pprof.WriteHeapProfile(p.memProfile)
		p.memProfile.Close()
		p.memProfile = nil
	}
	
	if p.blockProfile != nil {
		pprof.Lookup("block").WriteTo(p.blockProfile, 0)
		p.blockProfile.Close()
		p.blockProfile = nil
		runtime.SetBlockProfileRate(0)
	}
	
	if p.mutexProfile != nil {
		pprof.Lookup("mutex").WriteTo(p.mutexProfile, 0)
		p.mutexProfile.Close()
		p.mutexProfile = nil
		runtime.SetMutexProfileFraction(0)
	}
	
	return nil
}

func (p *Profiler) GetStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"threads":        runtime.GOMAXPROCS(0),
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_idle":      m.HeapIdle,
		"heap_inuse":     m.HeapInuse,
		"heap_released":  m.HeapReleased,
		"heap_objects":   m.HeapObjects,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"mspan_inuse":    m.MSpanInuse,
		"mspan_sys":      m.MSpanSys,
		"mcache_inuse":   m.MCacheInuse,
		"mcache_sys":     m.MCacheSys,
		"buck_hash_sys":  m.BuckHashSys,
		"gc_sys":         m.GCSys,
		"other_sys":      m.OtherSys,
		"next_gc":        m.NextGC,
		"last_gc":        m.LastGC,
		"pause_total_ns": m.PauseTotalNs,
		"pause_ns":       m.PauseNs[(m.NumGC+255)%256],
		"num_gc":         m.NumGC,
		"num_forced_gc":  m.NumForcedGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"enable_gc":      m.EnableGC,
		"debug_gc":       m.DebugGC,
	}
}

type PProfServer struct {
	addr   string
	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPProfServer(addr string) *PProfServer {
	if addr == "" {
		addr = ":6060"
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &PProfServer{
		addr:   addr,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ps *PProfServer) Start() error {
	ps.server = &http.Server{
		Addr:    ps.addr,
	}
	fmt.Printf("PProf server started on %s\n", ps.addr)
	fmt.Printf("Visit http://%s/debug/pprof/ for profiling data\n", ps.addr)
	return ps.server.ListenAndServe()
}

func (ps *PProfServer) Stop() error {
	ps.cancel()
	return ps.server.Shutdown(context.Background())
}

func (ps *PProfServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"threads":    runtime.GOMAXPROCS(0),
		"memory":     getMemoryStats(),
		"gc":         getGCStats(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (ps *PProfServer) handleGC(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		runtime.GC()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GC triggered"))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_idle":      m.HeapIdle,
		"heap_inuse":     m.HeapInuse,
		"heap_released":  m.HeapReleased,
		"heap_objects":   m.HeapObjects,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"mspan_inuse":    m.MSpanInuse,
		"mspan_sys":      m.MSpanSys,
		"mcache_inuse":   m.MCacheInuse,
		"mcache_sys":     m.MCacheSys,
		"buck_hash_sys":  m.BuckHashSys,
		"gc_sys":         m.GCSys,
		"other_sys":      m.OtherSys,
	}
}

func getGCStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"next_gc":         m.NextGC,
		"last_gc":         m.LastGC,
		"pause_total_ns":  m.PauseTotalNs,
		"pause_ns":        m.PauseNs[(m.NumGC+255)%256],
		"num_gc":          m.NumGC,
		"num_forced_gc":   m.NumForcedGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"enable_gc":       m.EnableGC,
		"debug_gc":        m.DebugGC,
	}
}

type PerformanceAnalyzer struct {
	ctx           context.Context
	cancel        context.CancelFunc
	profiler      *Profiler
	metrics       map[string]interface{}
	mu            sync.RWMutex
	
	analysisInterval time.Duration
	alertThresholds  map[string]float64
}

func NewPerformanceAnalyzer(ctx context.Context) *PerformanceAnalyzer {
	ctx, cancel := context.WithCancel(ctx)
	
	return &PerformanceAnalyzer{
		ctx:              ctx,
		cancel:           cancel,
		metrics:          make(map[string]interface{}),
		analysisInterval: 10 * time.Second,
		alertThresholds: map[string]float64{
			"memory_usage_mb": 100.0,
			"goroutine_count":  1000.0,
			"gc_fraction":      0.1,
		},
	}
}

func (pa *PerformanceAnalyzer) Start() {
	go pa.analyzePerformance()
}

func (pa *PerformanceAnalyzer) Stop() {
	pa.cancel()
}

func (pa *PerformanceAnalyzer) analyzePerformance() {
	ticker := time.NewTicker(pa.analysisInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pa.performAnalysis()
		case <-pa.ctx.Done():
			return
		}
	}
}

func (pa *PerformanceAnalyzer) performAnalysis() {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	
	stats := pa.profiler.GetStats()
	
	memoryMB := float64(stats["heap_sys"].(uint64)) / 1024 / 1024
	if memoryMB > pa.alertThresholds["memory_usage_mb"] {
		fmt.Printf("WARNING: High memory usage: %.2f MB\n", memoryMB)
	}
	
	goroutines := stats["goroutines"].(int)
	if float64(goroutines) > pa.alertThresholds["goroutine_count"] {
		fmt.Printf("WARNING: High goroutine count: %d\n", goroutines)
	}
	
	gcFraction := stats["gc_cpu_fraction"].(float64)
	if gcFraction > pa.alertThresholds["gc_fraction"] {
		fmt.Printf("WARNING: High GC CPU fraction: %.2f%%\n", gcFraction*100)
	}
	
	pa.metrics[time.Now().Format("2006-01-02T15:04:05")] = stats
}

func (pa *PerformanceAnalyzer) GetAnalysis() map[string]interface{} {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	
	return map[string]interface{}{
		"current_metrics": pa.metrics,
		"thresholds":      pa.alertThresholds,
		"analysis_time":   time.Now(),
	}
} 