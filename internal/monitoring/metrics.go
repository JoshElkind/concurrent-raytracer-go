package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type RenderMetrics struct {
	RaysPerSecond   int64
	PixelsPerSecond int64
	ActiveWorkers   int32
	CompletedJobs   int64
	TotalJobs       int64
	
	MemoryUsage     int64
	HeapAlloc       uint64
	HeapSys         uint64
	HeapIdle        uint64
	HeapInuse       uint64
	
	CPUUsage        float64
	GoroutineCount  int
	ThreadCount     int
	
	StartTime       time.Time
	ElapsedTime     time.Duration
	AverageJobTime  time.Duration
	
	AntiAliasingSamples int
	MaxRayDepth         int
	RecursiveReflections bool
	SoftShadows         bool
	
	ProgressPercent     float64
	EstimatedTimeRemaining time.Duration
	RatePerSecond       float64
}

type MetricsCollector struct {
	ctx           context.Context
	cancel        context.CancelFunc
	metrics       *RenderMetrics
	mu            sync.RWMutex
	
	rayCountChan    chan int64
	pixelCountChan  chan int64
	jobCompleteChan chan time.Duration
	
	observers []MetricsObserver
	
	collectionInterval time.Duration
}

type MetricsObserver interface {
	OnMetricsUpdate(metrics *RenderMetrics)
}

type ProgressReporter struct {
	metrics       *RenderMetrics
	startTime     time.Time
	totalPixels   int64
	completedPixels int64
	reportInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewMetricsCollector(ctx context.Context) *MetricsCollector {
	ctx, cancel := context.WithCancel(ctx)
	
	collector := &MetricsCollector{
		ctx:                 ctx,
		cancel:              cancel,
		metrics:             &RenderMetrics{StartTime: time.Now()},
		rayCountChan:        make(chan int64, 1000),
		pixelCountChan:      make(chan int64, 1000),
		jobCompleteChan:     make(chan time.Duration, 1000),
		observers:           make([]MetricsObserver, 0),
		collectionInterval:  time.Second,
	}
	
	return collector
}

func (mc *MetricsCollector) Start() {
	go mc.collectMetrics()
	go mc.updateSystemMetrics()
}

func (mc *MetricsCollector) Stop() {
	mc.cancel()
}

func (mc *MetricsCollector) collectMetrics() {
	ticker := time.NewTicker(mc.collectionInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.updateMetrics()
		case <-mc.ctx.Done():
			return
		}
	}
}

func (mc *MetricsCollector) updateSystemMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.updateSystemStats()
		case <-mc.ctx.Done():
			return
		}
	}
}

func (mc *MetricsCollector) updateMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.ElapsedTime = time.Since(mc.metrics.StartTime)
	
	if mc.metrics.ElapsedTime > 0 {
		mc.metrics.RatePerSecond = float64(mc.metrics.CompletedJobs) / mc.metrics.ElapsedTime.Seconds()
	}
	
	if mc.metrics.TotalJobs > 0 {
		mc.metrics.ProgressPercent = float64(mc.metrics.CompletedJobs) / float64(mc.metrics.TotalJobs) * 100
	}
	
	if mc.metrics.RatePerSecond > 0 {
		remainingJobs := mc.metrics.TotalJobs - mc.metrics.CompletedJobs
		secondsRemaining := float64(remainingJobs) / mc.metrics.RatePerSecond
		mc.metrics.EstimatedTimeRemaining = time.Duration(secondsRemaining) * time.Second
	}
	
	for _, observer := range mc.observers {
		observer.OnMetricsUpdate(mc.metrics)
	}
}

func (mc *MetricsCollector) updateSystemStats() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	mc.metrics.HeapAlloc = m.HeapAlloc
	mc.metrics.HeapSys = m.HeapSys
	mc.metrics.HeapIdle = m.HeapIdle
	mc.metrics.HeapInuse = m.HeapInuse
	mc.metrics.MemoryUsage = int64(m.HeapAlloc)
	
	mc.metrics.GoroutineCount = runtime.NumGoroutine()
	
	mc.metrics.ThreadCount = runtime.GOMAXPROCS(0)
}

func (mc *MetricsCollector) RecordRay() {
	atomic.AddInt64(&mc.metrics.RaysPerSecond, 1)
	select {
	case mc.rayCountChan <- 1:
	default:
	}
}

func (mc *MetricsCollector) RecordPixel() {
	atomic.AddInt64(&mc.metrics.PixelsPerSecond, 1)
	select {
	case mc.pixelCountChan <- 1:
	default:
	}
}

func (mc *MetricsCollector) RecordJobComplete(duration time.Duration) {
	atomic.AddInt64(&mc.metrics.CompletedJobs, 1)
	select {
	case mc.jobCompleteChan <- duration:
	default:
	}
}

func (mc *MetricsCollector) SetActiveWorkers(count int32) {
	atomic.StoreInt32(&mc.metrics.ActiveWorkers, count)
}

func (mc *MetricsCollector) SetTotalJobs(count int64) {
	atomic.StoreInt64(&mc.metrics.TotalJobs, count)
}

func (mc *MetricsCollector) SetQualitySettings(samples, maxDepth int, reflections, shadows bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.AntiAliasingSamples = samples
	mc.metrics.MaxRayDepth = maxDepth
	mc.metrics.RecursiveReflections = reflections
	mc.metrics.SoftShadows = shadows
}

func (mc *MetricsCollector) AddObserver(observer MetricsObserver) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.observers = append(mc.observers, observer)
}

func (mc *MetricsCollector) GetMetrics() *RenderMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	metrics := *mc.metrics
	return &metrics
}

func NewProgressReporter(ctx context.Context, totalPixels int64) *ProgressReporter {
	ctx, cancel := context.WithCancel(ctx)
	
	return &ProgressReporter{
		metrics:        &RenderMetrics{},
		startTime:      time.Now(),
		totalPixels:    totalPixels,
		reportInterval: 100 * time.Millisecond,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (pr *ProgressReporter) Start() {
	go pr.reportProgress()
}

func (pr *ProgressReporter) Stop() {
	pr.cancel()
}

func (pr *ProgressReporter) UpdateProgress(completedPixels int64) {
	atomic.StoreInt64(&pr.completedPixels, completedPixels)
}

func (pr *ProgressReporter) reportProgress() {
	ticker := time.NewTicker(pr.reportInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pr.printProgress()
		case <-pr.ctx.Done():
			return
		}
	}
}

func (pr *ProgressReporter) printProgress() {
	completed := atomic.LoadInt64(&pr.completedPixels)
	elapsed := time.Since(pr.startTime)
	
	if elapsed > 0 && pr.totalPixels > 0 {
		progress := float64(completed) / float64(pr.totalPixels) * 100
		rate := float64(completed) / elapsed.Seconds()
		
		eta := time.Duration(0)
		if rate > 0 {
			remaining := pr.totalPixels - completed
			eta = time.Duration(float64(remaining)/rate) * time.Second
		}
		
		pr.printProgressBar(progress, rate, eta)
	}
}

func (pr *ProgressReporter) printProgressBar(progress, rate float64, eta time.Duration) {
	barWidth := 50
	filled := int(progress / 100 * float64(barWidth))
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"
	
	print("\r")
	fmt.Printf("%s %.1f%% | %.0f pixels/sec | ETA: %v", 
		bar, progress, rate, eta)
}

func (pr *ProgressReporter) estimateTimeRemaining() time.Duration {
	completed := atomic.LoadInt64(&pr.completedPixels)
	elapsed := time.Since(pr.startTime)
	
	if elapsed > 0 && completed > 0 {
		rate := float64(completed) / elapsed.Seconds()
		if rate > 0 {
			remaining := pr.totalPixels - completed
			return time.Duration(float64(remaining)/rate) * time.Second
		}
	}
	
	return 0
}

type PerformanceMonitor struct {
	metrics       *RenderMetrics
	alertChan     chan PerformanceAlert
	thresholds    PerformanceThresholds
	ctx           context.Context
	cancel        context.CancelFunc
}

type PerformanceAlert struct {
	Type      string
	Message   string
	Severity  string
	Timestamp time.Time
	Metrics   *RenderMetrics
}

type PerformanceThresholds struct {
	MaxMemoryUsage     int64
	MinRaysPerSecond   int64
	MaxJobTime         time.Duration
	MinCPUUtilization  float64
}

func NewPerformanceMonitor(ctx context.Context) *PerformanceMonitor {
	ctx, cancel := context.WithCancel(ctx)
	
	return &PerformanceMonitor{
		metrics:    &RenderMetrics{},
		alertChan:  make(chan PerformanceAlert, 100),
		thresholds: PerformanceThresholds{
			MaxMemoryUsage:   100 * 1024 * 1024, // 100MB
			MinRaysPerSecond: 1000,
			MaxJobTime:       1 * time.Second,
			MinCPUUtilization: 70.0,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (pm *PerformanceMonitor) Start() {
	go pm.monitorPerformance()
}

func (pm *PerformanceMonitor) Stop() {
	pm.cancel()
}

func (pm *PerformanceMonitor) monitorPerformance() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pm.checkThresholds()
		case <-pm.ctx.Done():
			return
		}
	}
}

func (pm *PerformanceMonitor) checkThresholds() {
	if pm.metrics.MemoryUsage > pm.thresholds.MaxMemoryUsage {
		pm.sendAlert("MEMORY_HIGH", "Memory usage exceeded threshold", "WARNING")
	}
	
	if pm.metrics.RaysPerSecond < pm.thresholds.MinRaysPerSecond {
		pm.sendAlert("PERFORMANCE_LOW", "Ray processing rate below threshold", "WARNING")
	}
	
	if pm.metrics.CPUUsage < pm.thresholds.MinCPUUtilization {
		pm.sendAlert("CPU_LOW", "CPU utilization below threshold", "INFO")
	}
}

func (pm *PerformanceMonitor) sendAlert(alertType, message, severity string) {
	alert := PerformanceAlert{
		Type:      alertType,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
		Metrics:   pm.metrics,
	}
	
	select {
	case pm.alertChan <- alert:
	default:
	}
}

func (pm *PerformanceMonitor) GetAlerts() <-chan PerformanceAlert {
	return pm.alertChan
} 