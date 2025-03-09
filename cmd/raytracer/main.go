package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"raytraceGo/internal/renderer"
	"raytraceGo/internal/scene"
	"runtime"
	"strconv"
)

func main() {
	flag.Parse()
	args := flag.Args()
	
	if len(args) < 4 {
		fmt.Println("Usage: raytracer <scene_file> <output_file> <width> <height>")
		fmt.Println("Example: raytracer scene.json output.png 800 600")
		os.Exit(1)
	}
	
	sceneFile := args[0]
	outputFile := args[1]
	width, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Invalid width: %s\n", args[2])
		os.Exit(1)
	}
	
	height, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Printf("Invalid height: %s\n", args[3])
		os.Exit(1)
	}
	
	fmt.Printf("Loading scene from: %s\n", sceneFile)
	
	scene, err := scene.LoadFromFile(sceneFile)
	if err != nil {
		fmt.Printf("Error loading scene: %v\n", err)
		os.Exit(1)
	}
	
	numWorkers := runtime.NumCPU()
	renderer := renderer.NewParallelRenderer(numWorkers)
	
	fmt.Printf("Rendering at %dx%d resolution...\n", width, height)
	
	img := renderer.Render(scene, width, height)
	
	outputPath := outputFile
	if filepath.Ext(outputPath) == "" {
		outputPath += ".png"
	}
	
	fmt.Printf("Saving to: %s\n", outputPath)
	if err := renderer.SaveImage(img, outputPath); err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		os.Exit(1)
	}
	
	benchmarkPath := filepath.Join(filepath.Dir(outputPath), "benchmark_data.json")
	if err := renderer.SaveBenchmarkData(benchmarkPath); err != nil {
		fmt.Printf("Error saving benchmark data: %v\n", err)
	} else {
		fmt.Println("Benchmark data saved")
	}
} 
