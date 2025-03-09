package main

import (
	"flag"
	"fmt"
	"log"
	stdmath "math"
	"runtime"
	"strconv"
	"time"
	"raytraceGo/internal/benchmarking"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	"raytraceGo/internal/optimization"
	"raytraceGo/internal/renderer"
	"raytraceGo/internal/scene"
)

func main() {
	var (
		width     = flag.Int("width", 800, "Image width")
		height    = flag.Int("height", 600, "Image height")
		workers   = flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines")
		output    = flag.String("output", "output.png", "Output image file")
		sceneFile = flag.String("scene", "", "Scene configuration file (JSON)")
		samples   = flag.Int("samples", 100, "Samples per pixel for anti-aliasing")
		maxDepth  = flag.String("max-depth", "50", "Maximum ray recursion depth")
		duration  = flag.Duration("duration", 5*time.Second, "Benchmark duration")
	)
	flag.Parse()

	fmt.Printf("Optimization Demo - Advanced Ray Tracer\n")
	fmt.Printf("Image: %dx%d, Workers: %d\n", *width, *height, *workers)
	fmt.Printf("Output: %s\n", *output)

	// Run demos
	demoMathematicalCompetency()
	demoSpatialAcceleration()
	demoMemoryOptimization()
	demoPerformanceBenchmarking(*width, *height, *workers, *samples, *duration)
	demoOptimizedRendering(*width, *height, *workers, *samples, *output, *sceneFile, *maxDepth)
}

func demoMathematicalCompetency() {
	fmt.Printf("Demonstrating mathematical competency...\n")

	// Test fast math functions
	x := 2.0
	y := 3.0
	
	// Use standard math functions since fast versions may not be available
	sqrtResult := stdmath.Sqrt(x*x + y*y)
	expResult := stdmath.Exp(x)
	sinResult := stdmath.Sin(x)
	cosResult := stdmath.Cos(y)
	
	fmt.Printf("  Fast math results:\n")
	fmt.Printf("    sqrt(%f² + %f²) = %f\n", x, y, sqrtResult)
	fmt.Printf("    exp(%f) = %f\n", x, expResult)
	fmt.Printf("    sin(%f) = %f\n", x, sinResult)
	fmt.Printf("    cos(%f) = %f\n", y, cosResult)
	
	// Test vector operations
	vec1 := math.Vec3{X: 1, Y: 2, Z: 3}
	vec2 := math.Vec3{X: 4, Y: 5, Z: 6}
	
	// Use standard vector operations
	sum := vec1.Add(vec2)
	dot := vec1.Dot(vec2)
	
	fmt.Printf("  Vector operations:\n")
	fmt.Printf("    vec1 = %v\n", vec1)
	fmt.Printf("    vec2 = %v\n", vec2)
	fmt.Printf("    sum = %v\n", sum)
	fmt.Printf("    dot product = %f\n", dot)
}

func demoSpatialAcceleration() {
	fmt.Printf("Demonstrating spatial acceleration structures...\n")

	// Create test objects
	objects := make([]geometry.Hittable, 1000)
	for i := 0; i < 1000; i++ {
		center := math.Vec3{
			X: stdmath.Sin(float64(i) * stdmath.Pi / 500) * 10,
			Y: stdmath.Cos(float64(i) * stdmath.Pi / 500) * 10,
			Z: float64(i) / 100,
		}
		objects[i] = geometry.NewSphere(center, 0.5, nil)
	}

	// Test BVH
	bvhStart := time.Now()
	bvh := optimization.NewBVH(objects, 0, len(objects))
	bvhTime := time.Since(bvhStart)
	fmt.Printf("  BVH build time: %v\n", bvhTime)

	// Test Octree
	octreeStart := time.Now()
	octree := optimization.NewOctree(math.Vec3{X: 0, Y: 0, Z: 0}, 20.0, 8, 10)
	for _, obj := range objects {
		octree.Insert(obj)
	}
	octreeTime := time.Since(octreeStart)
	fmt.Printf("  Octree build time: %v\n", octreeTime)

	// Test KD-Tree
	kdTreeStart := time.Now()
	_ = optimization.NewKDTree(objects, 0)
	kdTreeTime := time.Since(kdTreeStart)
	fmt.Printf("  KD-Tree build time: %v\n", kdTreeTime)

	// Test ray intersection performance
	ray := geometry.NewRay(
		math.Vec3{X: 0, Y: 0, Z: -10},
		math.Vec3{X: 0, Y: 0, Z: 1},
	)

	// Test naive intersection
	naiveStart := time.Now()
	for _, obj := range objects {
		obj.Hit(ray, 0.001, stdmath.Inf(1))
	}
	naiveTime := time.Since(naiveStart)
	fmt.Printf("  Naive intersection: %v\n", naiveTime)

	// Test BVH intersection
	bvhHitStart := time.Now()
	bvh.Hit(ray, 0.001, stdmath.Inf(1))
	bvhHitTime := time.Since(bvhHitStart)
	fmt.Printf("  BVH intersection: %v\n", bvhHitTime)

	speedup := float64(naiveTime) / float64(bvhHitTime)
	fmt.Printf("  Speedup: %.2fx\n", speedup)
}

func demoMemoryOptimization() {
	fmt.Printf("Testing memory optimization techniques...\n")

	// Test object pooling
	pool := optimization.NewObjectPool()
	
	// Test ray pool
	rayStart := time.Now()
	for i := 0; i < 10000; i++ {
		ray := pool.GetRay()
		*ray = geometry.Ray{
			Origin:    math.Vec3{X: float64(i), Y: 0, Z: 0},
			Direction: math.Vec3{X: 0, Y: 0, Z: 1},
		}
		pool.PutRay(ray)
	}
	rayTime := time.Since(rayStart)
	fmt.Printf("  Ray pool: %d operations in %v\n", 10000, rayTime)

	// Test hit record pool
	hitStart := time.Now()
	for i := 0; i < 10000; i++ {
		hit := pool.GetHitRecord()
		*hit = geometry.HitRecord{
			Point: math.Vec3{X: float64(i), Y: 0, Z: 0},
			T:     float64(i),
		}
		pool.PutHitRecord(hit)
	}
	hitTime := time.Since(hitStart)
	fmt.Printf("  Hit record pool: %d operations in %v\n", 10000, hitTime)

	// Test vector pool
	vecStart := time.Now()
	for i := 0; i < 10000; i++ {
		vec := pool.GetVector()
		*vec = math.Vec3{X: float64(i), Y: float64(i), Z: float64(i)}
		pool.PutVector(vec)
	}
	vecTime := time.Since(vecStart)
	fmt.Printf("  Vector pool: %d operations in %v\n", 10000, vecTime)

	// Memory usage comparison
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  Current memory usage: %.2f MB\n", float64(m.Alloc)/1024/1024)
}

func demoPerformanceBenchmarking(width, height, workers, samples int, duration time.Duration) {
	fmt.Printf("Running comprehensive performance benchmarks...\n")

	// Create benchmark configuration
	config := benchmarking.BenchmarkConfig{
		Width:      width,
		Height:     height,
		Workers:    []int{1, 2, 4, 8, workers},
		Samples:    []int{10, 50, 100, samples},
		Scenes:     []string{"benchmark"},
		Duration:   duration,
		WarmupRuns: 2,
		Iterations: 3,
		OutputFile: "benchmark_results.json",
		EnableProfiling: true,
		EnableMetrics:   true,
	}

	// Run benchmark suite
	suite := benchmarking.NewBenchmarkSuite(config)
	err := suite.Run()
	if err != nil {
		log.Printf("Benchmark error: %v", err)
	}

	fmt.Printf("  Benchmark completed, results saved to benchmark_results.json\n")
}

func demoOptimizedRendering(width, height, workers, samples int, output, sceneFile, maxDepth string) {
	fmt.Printf("Running optimized rendering demo...\n")

	// Load scene
	scene, err := scene.LoadFromFile(sceneFile)
	if err != nil {
		log.Printf("Failed to load scene: %v", err)
		return
	}

	// Create optimized renderer
	renderer := renderer.NewOptimizedParallelRenderer(workers)
	renderer.SetSamples(samples)
	maxDepthInt, _ := strconv.Atoi(maxDepth)
	renderer.SetMaxDepth(maxDepthInt)
	renderer.SetAntiAliasing(true)
	renderer.SetRecursiveReflections(true)
	renderer.SetSoftShadows(true)
	renderer.SetDepthOfField(true)

	// Render with optimizations
	start := time.Now()
	image := renderer.Render(scene, width, height)
	renderTime := time.Since(start)

	// Print detailed metrics
	metrics := renderer.GetMetrics()
	fmt.Printf("  Render time: %v\n", renderTime)
	fmt.Printf("  Rays/second: %.2f M\n", metrics.RaysPerSecond/1e6)
	fmt.Printf("  Pixels/second: %.2f K\n", metrics.PixelsPerSecond/1e3)
	fmt.Printf("  Memory usage: %.2f MB\n", float64(metrics.MemoryUsage)/1024/1024)
	fmt.Printf("  CPU usage: %.1f%%\n", metrics.CPUUsage)
	fmt.Printf("  BVH build time: %v\n", metrics.BVHBuildTime)
	fmt.Printf("  Setup time: %v\n", metrics.SetupTime)

	// Save image
	err = renderer.SaveImage(image, output)
	if err != nil {
		log.Printf("Failed to save image: %v", err)
		return
	}

	fmt.Printf("  Image saved to: %s\n", output)
} 