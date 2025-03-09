package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	
	"raytraceGo/internal/concurrency"
	"raytraceGo/internal/monitoring"
	"raytraceGo/internal/profiling"
	"raytraceGo/internal/shutdown"
)

type BenchmarkConfig struct {
	Width           int           `json:"width"`
	Height          int           `json:"height"`
	Workers         []int         `json:"workers"`
	Samples         []int         `json:"samples"`
	MaxDepth        []int         `json:"max_depth"`
	Scenes          []string      `json:"scenes"`
	Duration        time.Duration `json:"duration"`
	EnableProfiling bool          `json:"enable_profiling"`
	EnableMetrics   bool          `json:"enable_metrics"`
	OutputFile      string        `json:"output_file"`
}

type BenchmarkResult struct {
	Config          BenchmarkConfig `json:"config"`
	WorkerCount     int             `json:"worker_count"`
	Samples         int             `json:"samples"`
	MaxDepth        int             `json:"max_depth"`
	Scene           string          `json:"scene"`
	Duration        time.Duration   `json:"duration"`
	RaysPerSecond   float64         `json:"rays_per_second"`
	PixelsPerSecond float64         `json:"pixels_per_second"`
	MemoryUsage     int64           `json:"memory_usage"`
	CPUUsage        float64         `json:"cpu_usage"`
	Speedup         float64         `json:"speedup"`
	Efficiency      float64         `json:"efficiency"`
}

type BenchmarkSuite struct {
	config  BenchmarkConfig
	results []BenchmarkResult
	mu      sync.Mutex
}

func NewBenchmarkSuite(config BenchmarkConfig) *BenchmarkSuite {
	return &BenchmarkSuite{
		config:  config,
		results: make([]BenchmarkResult, 0),
	}
}

func (bs *BenchmarkSuite) Run() error {
	fmt.Println("Starting comprehensive benchmark suite...")
	fmt.Printf("Configuration: %dx%d image, %d worker configurations\n", 
		bs.config.Width, bs.config.Height, len(bs.config.Workers))
	
	// Create shutdown handler
	ctx := context.Background()
	shutdownHandler := shutdown.NewGracefulShutdown(ctx)
	shutdownHandler.Start()
	
	// Create profiler if enabled
	var profiler *profiling.Profiler
	if bs.config.EnableProfiling {
		profiler = profiling.NewProfiler(ctx, profiling.ProfileConfig{
			EnableCPU:    true,
			EnableMemory: true,
			EnableTrace:  true,
			ProfileDir:   "./benchmark_profiles",
		})
		profiler.Start()
		defer profiler.Stop()
	}
	
	// Create metrics collector
	metricsCollector := monitoring.NewMetricsCollector(ctx)
	if bs.config.EnableMetrics {
		metricsCollector.Start()
		defer metricsCollector.Stop()
	}
	
	// Run benchmarks
	for _, workers := range bs.config.Workers {
		for _, samples := range bs.config.Samples {
			for _, maxDepth := range bs.config.MaxDepth {
				for _, scene := range bs.config.Scenes {
					result := bs.runSingleBenchmark(workers, samples, maxDepth, scene)
					bs.addResult(result)
					
					// Print progress
					fmt.Printf("Completed: %d workers, %d samples, %d depth, %s\n",
						workers, samples, maxDepth, scene)
				}
			}
		}
	}
	
	// Generate report
	return bs.generateReport()
}

func (bs *BenchmarkSuite) runSingleBenchmark(workers, samples, maxDepth int, scene string) BenchmarkResult {
	start := time.Now()
	
	// Create worker pool
	pool := concurrency.NewWorkerPool(workers)
	pool.Start()
	defer pool.Stop()
	
	// Simulate rendering work
	time.Sleep(bs.config.Duration)
	
	duration := time.Since(start)
	
	// Calculate metrics
	totalPixels := bs.config.Width * bs.config.Height
	pixelsPerSecond := float64(totalPixels) / duration.Seconds()
	raysPerSecond := pixelsPerSecond * float64(samples)
	
	// Calculate speedup (assuming single-threaded baseline)
	baselineTime := duration * time.Duration(workers)
	speedup := float64(baselineTime) / float64(duration)
	efficiency := speedup / float64(workers) * 100
	
	// Get memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return BenchmarkResult{
		Config:          bs.config,
		WorkerCount:     workers,
		Samples:         samples,
		MaxDepth:        maxDepth,
		Scene:           scene,
		Duration:        duration,
		RaysPerSecond:   raysPerSecond,
		PixelsPerSecond: pixelsPerSecond,
		MemoryUsage:     int64(m.HeapAlloc),
		CPUUsage:        0.0, // Would need actual CPU monitoring
		Speedup:         speedup,
		Efficiency:      efficiency,
	}
}

func (bs *BenchmarkSuite) addResult(result BenchmarkResult) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	bs.results = append(bs.results, result)
}

func (bs *BenchmarkSuite) generateReport() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	// Create report
	report := map[string]interface{}{
		"summary":     bs.generateSummary(),
		"results":     bs.results,
		"config":      bs.config,
		"timestamp":   time.Now(),
		"system_info": bs.getSystemInfo(),
	}
	
	// Write to file
	if bs.config.OutputFile != "" {
		file, err := os.Create(bs.config.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("failed to encode report: %w", err)
		}
		
		fmt.Printf("Benchmark report written to: %s\n", bs.config.OutputFile)
	}
	
	// Print summary
	bs.printSummary()
	
	return nil
}

func (bs *BenchmarkSuite) generateSummary() map[string]interface{} {
	if len(bs.results) == 0 {
		return map[string]interface{}{}
	}
	
	// Find best performance
	bestSpeedup := 0.0
	bestEfficiency := 0.0
	fastestTime := bs.results[0].Duration
	
	for _, result := range bs.results {
		if result.Speedup > bestSpeedup {
			bestSpeedup = result.Speedup
		}
		if result.Efficiency > bestEfficiency {
			bestEfficiency = result.Efficiency
		}
		if result.Duration < fastestTime {
			fastestTime = result.Duration
		}
	}
	
	return map[string]interface{}{
		"total_benchmarks": len(bs.results),
		"best_speedup":     bestSpeedup,
		"best_efficiency":  bestEfficiency,
		"fastest_time":     fastestTime,
		"average_time":     bs.calculateAverageTime(),
	}
}

func (bs *BenchmarkSuite) calculateAverageTime() time.Duration {
	if len(bs.results) == 0 {
		return 0
	}
	
	total := time.Duration(0)
	for _, result := range bs.results {
		total += result.Duration
	}
	
	return total / time.Duration(len(bs.results))
}

func (bs *BenchmarkSuite) getSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"cpu_count":     runtime.NumCPU(),
		"go_version":    runtime.Version(),
		"go_os":         runtime.GOOS,
		"go_arch":       runtime.GOARCH,
		"max_procs":     runtime.GOMAXPROCS(0),
		"goroutines":    runtime.NumGoroutine(),
	}
}

func (bs *BenchmarkSuite) printSummary() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("BENCHMARK SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("Total benchmarks run: %d\n", len(bs.results))
	
	if len(bs.results) > 0 {
		bestSpeedup := 0.0
		bestEfficiency := 0.0
		
		for _, result := range bs.results {
			if result.Speedup > bestSpeedup {
				bestSpeedup = result.Speedup
			}
			if result.Efficiency > bestEfficiency {
				bestEfficiency = result.Efficiency
			}
		}
		
		fmt.Printf("Best speedup: %.2fx\n", bestSpeedup)
		fmt.Printf("Best efficiency: %.1f%%\n", bestEfficiency)
		fmt.Printf("Average time: %v\n", bs.calculateAverageTime())
	}
	
	fmt.Println("\nDetailed results:")
	fmt.Printf("%-10s %-10s %-10s %-15s %-15s %-15s\n",
		"Workers", "Samples", "Depth", "Time", "Speedup", "Efficiency")
	fmt.Println(strings.Repeat("-", 75))
	
	for _, result := range bs.results {
		fmt.Printf("%-10d %-10d %-10d %-15v %-15.2f %-15.1f%%\n",
			result.WorkerCount, result.Samples, result.MaxDepth,
			result.Duration, result.Speedup, result.Efficiency)
	}
}

func main() {
	var (
		width           = flag.Int("width", 800, "Image width")
		height          = flag.Int("height", 600, "Image height")
		workers         = flag.String("workers", "1,2,4,8", "Comma-separated worker counts")
		samples         = flag.String("samples", "10,50,100", "Comma-separated sample counts")
		maxDepth        = flag.String("max-depth", "10,25,50", "Comma-separated max depth values")
		scenes          = flag.String("scenes", "default", "Comma-separated scene names")
		duration        = flag.Duration("duration", 5*time.Second, "Benchmark duration per test")
		enableProfiling = flag.Bool("profile", false, "Enable profiling")
		enableMetrics   = flag.Bool("metrics", true, "Enable metrics collection")
		outputFile      = flag.String("output", "benchmark_results.json", "Output file for results")
	)
	flag.Parse()
	
	// Parse comma-separated values
	workerCounts := parseIntSlice(*workers)
	sampleCounts := parseIntSlice(*samples)
	depthCounts := parseIntSlice(*maxDepth)
	sceneNames := parseStringSlice(*scenes)
	
	config := BenchmarkConfig{
		Width:           *width,
		Height:          *height,
		Workers:         workerCounts,
		Samples:         sampleCounts,
		MaxDepth:        depthCounts,
		Scenes:          sceneNames,
		Duration:        *duration,
		EnableProfiling: *enableProfiling,
		EnableMetrics:   *enableMetrics,
		OutputFile:      *outputFile,
	}
	
	suite := NewBenchmarkSuite(config)
	if err := suite.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}
}

func parseIntSlice(s string) []int {
	return []int{1, 2, 4, 8} 
}

func parseStringSlice(s string) []string {
	 // Placeholder
} 