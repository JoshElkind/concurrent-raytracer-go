package benchmarking

import (
	"encoding/json"
	"fmt"
	stdmath "math"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	"raytraceGo/internal/optimization"
)

type BenchmarkConfig struct {
	Width           int
	Height          int
	Workers         []int
	Samples         []int
	Scenes          []string
	Duration        time.Duration
	WarmupRuns      int
	Iterations      int
	OutputFile      string
	EnableProfiling bool
	EnableMetrics   bool
}

type BenchmarkResult struct {
	Config          BenchmarkConfig
	WorkerCount     int
	SampleCount     int
	Scene           string
	RaysPerSecond   float64
	PixelsPerSecond float64
	MemoryUsage     uint64
	CPUUsage        float64
	RenderTime      time.Duration
	SetupTime       time.Duration
	CleanupTime     time.Duration
	TotalTime       time.Duration
	Error           error
}

type PerformanceMetrics struct {
	MinRaysPerSecond   float64
	MaxRaysPerSecond   float64
	AvgRaysPerSecond   float64
	MedianRaysPerSecond float64
	StdDevRaysPerSecond float64
	MinMemoryUsage     uint64
	MaxMemoryUsage     uint64
	AvgMemoryUsage     uint64
	MinCPUUsage        float64
	MaxCPUUsage        float64
	AvgCPUUsage        float64
}

type BenchmarkSuite struct {
	config    BenchmarkConfig
	results   []BenchmarkResult
	metrics   map[string]PerformanceMetrics
	mutex     sync.RWMutex
	startTime time.Time
}

func NewBenchmarkSuite(config BenchmarkConfig) *BenchmarkSuite {
	return &BenchmarkSuite{
		config:  config,
		results: make([]BenchmarkResult, 0),
		metrics: make(map[string]PerformanceMetrics),
	}
}

func (bs *BenchmarkSuite) Run() error {
	bs.startTime = time.Now()
	fmt.Printf("Starting comprehensive benchmark suite...\n")
	fmt.Printf("Configuration: %dx%d, %d workers, %d samples\n", 
		bs.config.Width, bs.config.Height, 
		bs.config.Workers[0], bs.config.Samples[0])
	
	if bs.config.WarmupRuns > 0 {
		fmt.Printf("Running %d warmup runs...\n", bs.config.WarmupRuns)
		bs.runWarmup()
	}
	
	totalRuns := len(bs.config.Workers) * len(bs.config.Samples) * len(bs.config.Scenes)
	currentRun := 0
	
	for _, workers := range bs.config.Workers {
		for _, samples := range bs.config.Samples {
			for _, scene := range bs.config.Scenes {
				currentRun++
				fmt.Printf("Progress: %d/%d (%.1f%%)\n", 
					currentRun, totalRuns, 
					float64(currentRun)/float64(totalRuns)*100)
				
				result := bs.runSingleBenchmark(workers, samples, scene)
				bs.addResult(result)
			}
		}
	}
	
	bs.calculateMetrics()
	
	return bs.generateReport()
}

func (bs *BenchmarkSuite) runWarmup() {
	for i := 0; i < bs.config.WarmupRuns; i++ {
		bs.runSingleBenchmark(
			bs.config.Workers[0],
			bs.config.Samples[0],
			bs.config.Scenes[0],
		)
	}
}

func (bs *BenchmarkSuite) runSingleBenchmark(workers, samples int, scene string) BenchmarkResult {
	result := BenchmarkResult{
		Config:      bs.config,
		WorkerCount: workers,
		SampleCount: samples,
		Scene:       scene,
	}
	
	setupStart := time.Now()
	sceneObjects := bs.createTestScene(workers, samples)
	setupTime := time.Since(setupStart)
	result.SetupTime = setupTime
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryBefore := m.Alloc
	
	renderStart := time.Now()
	raysPerSecond := bs.benchmarkRendering(sceneObjects, workers, samples)
	renderTime := time.Since(renderStart)
	result.RenderTime = renderTime
	result.RaysPerSecond = raysPerSecond
	
	runtime.ReadMemStats(&m)
	memoryAfter := m.Alloc
	result.MemoryUsage = memoryAfter - memoryBefore
	
	result.CPUUsage = float64(runtime.NumCPU()) * 0.8 // Approximation
	
	totalPixels := bs.config.Width * bs.config.Height
	result.PixelsPerSecond = float64(totalPixels) / renderTime.Seconds()
	
	cleanupStart := time.Now()
	bs.cleanup(sceneObjects)
	result.CleanupTime = time.Since(cleanupStart)
	
	result.TotalTime = setupTime + renderTime + result.CleanupTime
	
	return result
}

func (bs *BenchmarkSuite) createTestScene(workers, samples int) []geometry.Hittable {
	objects := make([]geometry.Hittable, 0)
	
	ground := geometry.NewPlane(
		math.Vec3{X: 0, Y: -1, Z: 0},
		math.Vec3{X: 0, Y: 1, Z: 0},
		nil,
	)
	objects = append(objects, ground)
	
	for i := 0; i < 10; i++ {
		center := math.Vec3{
			X: stdmath.Sin(float64(i) * stdmath.Pi / 5) * 3,
			Y: 0.5,
			Z: stdmath.Cos(float64(i) * stdmath.Pi / 5) * 3,
		}
		
		sphere := geometry.NewSphere(center, 0.5, nil)
		objects = append(objects, sphere)
	}
	
	return objects
}

func (bs *BenchmarkSuite) benchmarkRendering(objects []geometry.Hittable, workers, samples int) float64 {
	start := time.Now()
	
	raysProcessed := int64(0)
	var wg sync.WaitGroup
	
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < samples; j++ {
				ray := geometry.NewRay(
					math.Vec3{X: 0, Y: 0, Z: 0},
					math.Vec3{X: stdmath.Sin(float64(j)), Y: 0, Z: stdmath.Cos(float64(j))},
				)
				
				for _, obj := range objects {
					if _, hit := obj.Hit(ray, 0.001, stdmath.Inf(1)); hit {
						atomic.AddInt64(&raysProcessed, 1)
					}
				}
			}
		}()
	}
	
	wg.Wait()
	
	duration := time.Since(start)
	return float64(raysProcessed) / duration.Seconds()
}

func (bs *BenchmarkSuite) cleanup(objects []geometry.Hittable) {
}

func (bs *BenchmarkSuite) addResult(result BenchmarkResult) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.results = append(bs.results, result)
}

func (bs *BenchmarkSuite) calculateMetrics() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	grouped := make(map[string][]BenchmarkResult)
	for _, result := range bs.results {
		key := fmt.Sprintf("%d_%d_%s", result.WorkerCount, result.SampleCount, result.Scene)
		grouped[key] = append(grouped[key], result)
	}
	
	for key, results := range grouped {
		metrics := PerformanceMetrics{}
		
		raysPerSecond := make([]float64, len(results))
		for i, result := range results {
			raysPerSecond[i] = result.RaysPerSecond
		}
		sort.Float64s(raysPerSecond)
		
		metrics.MinRaysPerSecond = raysPerSecond[0]
		metrics.MaxRaysPerSecond = raysPerSecond[len(raysPerSecond)-1]
		metrics.MedianRaysPerSecond = raysPerSecond[len(raysPerSecond)/2]
		
		sum := 0.0
		for _, rps := range raysPerSecond {
			sum += rps
		}
		metrics.AvgRaysPerSecond = sum / float64(len(raysPerSecond))
		
		variance := 0.0
		for _, rps := range raysPerSecond {
			diff := rps - metrics.AvgRaysPerSecond
			variance += diff * diff
		}
		metrics.StdDevRaysPerSecond = stdmath.Sqrt(variance / float64(len(raysPerSecond)))
		
		memoryUsage := make([]uint64, len(results))
		for i, result := range results {
			memoryUsage[i] = result.MemoryUsage
		}
		sort.Slice(memoryUsage, func(i, j int) bool {
			return memoryUsage[i] < memoryUsage[j]
		})
		
		metrics.MinMemoryUsage = memoryUsage[0]
		metrics.MaxMemoryUsage = memoryUsage[len(memoryUsage)-1]
		
		sumMem := uint64(0)
		for _, mem := range memoryUsage {
			sumMem += mem
		}
		metrics.AvgMemoryUsage = sumMem / uint64(len(memoryUsage))
		
		cpuUsage := make([]float64, len(results))
		for i, result := range results {
			cpuUsage[i] = result.CPUUsage
		}
		sort.Float64s(cpuUsage)
		
		metrics.MinCPUUsage = cpuUsage[0]
		metrics.MaxCPUUsage = cpuUsage[len(cpuUsage)-1]
		
		sumCPU := 0.0
		for _, cpu := range cpuUsage {
			sumCPU += cpu
		}
		metrics.AvgCPUUsage = sumCPU / float64(len(cpuUsage))
		
		bs.metrics[key] = metrics
	}
}

func (bs *BenchmarkSuite) generateReport() error {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	report := struct {
		Config    BenchmarkConfig
		Results   []BenchmarkResult
		Metrics   map[string]PerformanceMetrics
		Summary   string
		Timestamp time.Time
	}{
		Config:    bs.config,
		Results:   bs.results,
		Metrics:   bs.metrics,
		Summary:   bs.generateSummary(),
		Timestamp: time.Now(),
	}
	
	if bs.config.OutputFile != "" {
		file, err := os.Create(bs.config.OutputFile)
		if err != nil {
			return err
		}
		defer file.Close()
		
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	
	fmt.Println(bs.generateSummary())
	return nil
}

func (bs *BenchmarkSuite) generateSummary() string {
	summary := fmt.Sprintf(`
=== COMPREHENSIVE BENCHMARK SUMMARY ===
Total Duration: %v
Total Results: %d
Configuration: %dx%d pixels

PERFORMANCE SUMMARY:
`, time.Since(bs.startTime), len(bs.results), bs.config.Width, bs.config.Height)
	
	bestRaysPerSecond := 0.0
	bestConfig := ""
	
	for key, metrics := range bs.metrics {
		if metrics.AvgRaysPerSecond > bestRaysPerSecond {
			bestRaysPerSecond = metrics.AvgRaysPerSecond
			bestConfig = key
		}
	}
	
	summary += fmt.Sprintf("Best Performance: %.2f rays/sec (%s)\n", bestRaysPerSecond, bestConfig)
	
	totalMemory := uint64(0)
	for _, result := range bs.results {
		totalMemory += result.MemoryUsage
	}
	avgMemory := totalMemory / uint64(len(bs.results))
	
	summary += fmt.Sprintf("Average Memory Usage: %d bytes (%.2f MB)\n", 
		avgMemory, float64(avgMemory)/1024/1024)
	
	totalCPU := 0.0
	for _, result := range bs.results {
		totalCPU += result.CPUUsage
	}
	avgCPU := totalCPU / float64(len(bs.results))
	
	summary += fmt.Sprintf("Average CPU Usage: %.1f%%\n", avgCPU)
	
	summary += "\nSCALING ANALYSIS:\n"
	workerScaling := make(map[int][]float64)
	for _, result := range bs.results {
		workerScaling[result.WorkerCount] = append(workerScaling[result.WorkerCount], result.RaysPerSecond)
	}
	
	for workers, performances := range workerScaling {
		avg := 0.0
		for _, perf := range performances {
			avg += perf
		}
		avg /= float64(len(performances))
		
		efficiency := avg / float64(workers) / (avg / float64(1)) * 100
		summary += fmt.Sprintf("  %d workers: %.2f rays/sec (%.1f%% efficiency)\n", 
			workers, avg, efficiency)
	}
	
	return summary
}

func QuickBenchmark(width, height, workers, samples int) BenchmarkResult {
	config := BenchmarkConfig{
		Width:      width,
		Height:     height,
		Workers:    []int{workers},
		Samples:    []int{samples},
		Scenes:     []string{"quick"},
		Duration:   5 * time.Second,
		WarmupRuns: 1,
		Iterations: 3,
	}
	
	suite := NewBenchmarkSuite(config)
	suite.Run()
	
	if len(suite.results) > 0 {
		return suite.results[0]
	}
	
	return BenchmarkResult{Error: fmt.Errorf("no benchmark results")}
}

func MemoryBenchmark(objects []geometry.Hittable) uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	baseline := m.Alloc
	
	_ = optimization.NewBVH(objects, 0, len(objects))
	
	runtime.ReadMemStats(&m)
	return m.Alloc - baseline
}

func CPUBenchmark(iterations int) float64 {
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		result := 0.0
		for j := 0; j < 1000000; j++ {
			result += stdmath.Sin(float64(j)) * stdmath.Cos(float64(j))
		}
		_ = result
	}
	
	duration := time.Since(start)
	return float64(iterations) / duration.Seconds()
} 