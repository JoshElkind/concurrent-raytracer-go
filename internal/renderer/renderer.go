package renderer

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	stdmath "math"
	"os"
	"path/filepath"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	"raytraceGo/internal/material"
	"raytraceGo/internal/scene"
	"sync"
	"time"
	"encoding/json"
)

type ParallelRenderer struct {
	numWorkers int
	maxDepth   int
	samples    int
	antiAliasing bool
	recursiveReflections bool
	softShadows bool
	depthOfField bool
	benchmarkData *BenchmarkData
}

type BenchmarkData struct {
	SceneName     string    `json:"scene_name"`
	Resolution    string    `json:"resolution"`
	RenderTime    float64   `json:"render_time_seconds"`
	Samples       int       `json:"samples"`
	MaxDepth      int       `json:"max_depth"`
	NumWorkers    int       `json:"num_workers"`
	Objects       int       `json:"objects"`
	Lights        int       `json:"lights"`
	Timestamp     time.Time `json:"timestamp"`
	Features      []string  `json:"features"`
}

type RenderResult struct {
	pixels []Pixel
	startX, startY int
}

type Pixel struct {
	x, y int
	color math.Vec3
}

func NewParallelRenderer(numWorkers int) *ParallelRenderer {
	return &ParallelRenderer{
		numWorkers:           numWorkers,
		maxDepth:             50,
		samples:              100,
		antiAliasing:         true,
		recursiveReflections: true,
		softShadows:          true,
		depthOfField:         false,
		benchmarkData:        &BenchmarkData{},
	}
}

func (r *ParallelRenderer) Render(scene *scene.Scene, width, height int) *image.RGBA {
	startTime := time.Now()
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	camera := r.setupCamera(scene.Camera, width, height)
	hittables := scene.GetHittables()
	lights := scene.GetLights()
	
	tasks := r.createRenderTasks(width, height, scene, camera)
	results := make(chan RenderResult, r.numWorkers*2)
	
	var wg sync.WaitGroup
	
	for i := 0; i < r.numWorkers; i++ {
		wg.Add(1)
		go r.worker(&wg, tasks, results, hittables, lights)
	}
	
	go func() {
		wg.Wait()
		close(results)
	}()
	
	resultCount := 0
	for result := range results {
		for _, pixel := range result.pixels {
			mappedColor := r.toneMap(pixel.color)
			r, g, b := mappedColor.ToRGB()
			img.Set(pixel.x, pixel.y, color.RGBA{uint8(r), uint8(g), uint8(b), 255})
		}
		resultCount++
	}
	
	renderTime := time.Since(startTime).Seconds()
	
	r.benchmarkData.SceneName = scene.GetSceneName()
	r.benchmarkData.Resolution = fmt.Sprintf("%dx%d", width, height)
	r.benchmarkData.RenderTime = renderTime
	r.benchmarkData.Samples = r.samples
	r.benchmarkData.MaxDepth = r.maxDepth
	r.benchmarkData.NumWorkers = r.numWorkers
	r.benchmarkData.Objects = len(hittables)
	r.benchmarkData.Lights = len(lights)
	r.benchmarkData.Timestamp = time.Now()
	r.benchmarkData.Features = []string{
		"Improved metallic reflections with Fresnel effect",
		"Shiny materials with configurable roughness and specular",
		"Enhanced light source reflections",
		"Better specular highlights for metallic surfaces",
	}
	
	fmt.Printf("Rendering complete!\n")
	fmt.Printf("Enhanced materials features:\n")
	for _, feature := range r.benchmarkData.Features {
		fmt.Printf("- %s\n", feature)
	}
	
	return img
}

func (r *ParallelRenderer) worker(wg *sync.WaitGroup, tasks chan RenderTask, results chan RenderResult, hittables []geometry.Hittable, lights []scene.Light) {
	defer wg.Done()
	
	for task := range tasks {
		pixels := r.renderTile(task, hittables, lights)
		results <- RenderResult{pixels: pixels, startX: task.startX, startY: task.startY}
	}
}

func (r *ParallelRenderer) renderTile(task RenderTask, hittables []geometry.Hittable, lights []scene.Light) []Pixel {
	var pixels []Pixel
	
	for y := task.startY; y < task.endY; y++ {
		for x := task.startX; x < task.endX; x++ {
			color := r.tracePixel(x, y, task.width, task.height, task.camera, hittables, lights)
			pixels = append(pixels, Pixel{x: x, y: y, color: color})
		}
	}
	
	return pixels
}

func (r *ParallelRenderer) tracePixel(x, y, width, height int, camera *scene.Camera, hittables []geometry.Hittable, lights []scene.Light) math.Vec3 {
	color := math.Vec3{}
	samples := r.samples
	
	for s := 0; s < samples; s++ {
		u := (float64(x) + math.RandomFloat()) / float64(width)
		v := (float64(y) + math.RandomFloat()) / float64(height)
		
		ray := r.getRay(u, v, camera)
		color = color.Add(r.traceRay(ray, hittables, lights, 0))
	}
	
	return color.DivScalar(float64(samples))
}

func (r *ParallelRenderer) traceRay(ray geometry.Ray, hittables []geometry.Hittable, lights []scene.Light, depth int) math.Vec3 {
	if depth >= r.maxDepth {
		return math.Vec3{}
	}
	
	hitRecord, hit := r.hitWorld(ray, hittables, 0.001, stdmath.Inf(1))
	if !hit {
		return math.Vec3{X: 0.0, Y: 0.0, Z: 0.0}
	}
	
	material := hitRecord.Material.(material.Material)
	
	emitted := material.Emitted()
	
	directLighting := r.calculateDirectLighting(hitRecord, hittables, lights)
	
	scattered, attenuation, scatteredHit := material.Scatter(ray, hitRecord)
	if !scatteredHit {
		return emitted.Add(directLighting)
	}
	
	reflectedColor := math.Vec3{}
	if r.recursiveReflections {
		reflectedColor = r.traceRay(scattered, hittables, lights, depth+1)
	}
	
	metallic := material.GetMetallic()
	
	if metallic > 0.95 {
		reflectionWeight := 0.85
		directWeight := 0.15
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	} else if metallic > 0.9 {
		reflectionWeight := 0.8
		directWeight := 0.2
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	} else if metallic > 0.8 {
		reflectionWeight := 0.75
		directWeight := 0.25
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	} else if metallic > 0.7 {
		reflectionWeight := 0.7
		directWeight := 0.3
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	} else if metallic > 0.5 {
		reflectionWeight := 0.6
		directWeight := 0.4
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	} else if metallic > 0.2 {
		reflectionWeight := 0.4
		directWeight := 0.6
		finalColor := emitted.Add(directLighting.MulScalar(directWeight)).Add(attenuation.Mul(reflectedColor).MulScalar(reflectionWeight))
		return finalColor
	}
	
	finalColor := emitted.Add(directLighting).Add(attenuation.Mul(reflectedColor))
	return finalColor
}

func (r *ParallelRenderer) calculateDirectLighting(hit *geometry.HitRecord, hittables []geometry.Hittable, lights []scene.Light) math.Vec3 {
	totalLighting := math.Vec3{}
	
	material := hit.Material.(material.Material)
	albedo := material.GetAlbedo()
	metallic := material.GetMetallic()
	
	ambientStrength := 0.1
	if metallic > 0.9 {
		ambientStrength = 0.05
	} else if metallic > 0.7 {
		ambientStrength = 0.07
	} else if metallic > 0.5 {
		ambientStrength = 0.08
	}
	
	ambientLight := math.Vec3{X: ambientStrength, Y: ambientStrength, Z: ambientStrength}
	totalLighting = totalLighting.Add(ambientLight)
	
	for _, light := range lights {
		lightDir := light.Position.Sub(hit.Point).Normalize()
		lightDistance := light.Position.Sub(hit.Point).Length()
		
		if lightDistance < 0.001 {
			continue
		}
		
		shadowFactor := r.calculateSmartShadow(hit, light, hittables)
		
		if shadowFactor > 0.0 {
			cosTheta := stdmath.Max(0, hit.Normal.Dot(lightDir))
			intensity := cosTheta * light.Intensity / (lightDistance * lightDistance)
			
			diffuseStrength := 0.25
			if metallic > 0.95 {
				diffuseStrength = 0.05
			} else if metallic > 0.9 {
				diffuseStrength = 0.08
			} else if metallic > 0.8 {
				diffuseStrength = 0.12
			} else if metallic > 0.7 {
				diffuseStrength = 0.15
			} else if metallic > 0.5 {
				diffuseStrength = 0.2
			}
			
			diffuse := albedo.MulScalar(diffuseStrength * intensity * shadowFactor)
			totalLighting = totalLighting.Add(diffuse)
			
			if metallic > 0.5 {
				viewDir := hit.Point.MulScalar(-1).Normalize()
				halfDir := lightDir.Add(viewDir).Normalize()
				
				specularPower := 32.0
				if metallic > 0.9 {
					specularPower = 64.0
				} else if metallic > 0.8 {
					specularPower = 48.0
				}
				
				specularIntensity := stdmath.Pow(stdmath.Max(0, hit.Normal.Dot(halfDir)), specularPower)
				specular := light.Color.MulScalar(specularIntensity * intensity * shadowFactor * metallic * 3.0)
				totalLighting = totalLighting.Add(specular)
			}
		}
	}
	
	return totalLighting
}

func (r *ParallelRenderer) calculateSmartShadow(hit *geometry.HitRecord, light scene.Light, hittables []geometry.Hittable) float64 {
	lightDir := light.Position.Sub(hit.Point).Normalize()
	lightDistance := light.Position.Sub(hit.Point).Length()
	
	shadowRay := geometry.NewRay(hit.Point, lightDir)
	
	_, hitShadow := r.hitWorld(shadowRay, hittables, 0.001, lightDistance)
	
	if hitShadow {
		return 0.0
	}
	
	if r.softShadows {
		shadowSamples := 16
		shadowSum := 0.0
		
		for i := 0; i < shadowSamples; i++ {
			randomOffset := math.RandomVec3InUnitSphere().MulScalar(0.1)
			softLightDir := lightDir.Add(randomOffset).Normalize()
			softShadowRay := geometry.NewRay(hit.Point, softLightDir)
			
			_, softHit := r.hitWorld(softShadowRay, hittables, 0.001, lightDistance)
			
			if !softHit {
				shadowSum += 1.0
			}
		}
		
		return shadowSum / float64(shadowSamples)
	}
	
	return 1.0
}

func (r *ParallelRenderer) hitWorld(ray geometry.Ray, hittables []geometry.Hittable, tMin, tMax float64) (*geometry.HitRecord, bool) {
	var closestHit *geometry.HitRecord
	closestT := tMax
	
	for _, hittable := range hittables {
		hitRecord, hit := hittable.Hit(ray, tMin, closestT)
		if hit {
			closestT = hitRecord.T
			closestHit = hitRecord
		}
	}
	
	return closestHit, closestHit != nil
}

func (r *ParallelRenderer) toneMap(color math.Vec3) math.Vec3 {
	exposure := 1.0
	gamma := 2.2
	
	color = color.MulScalar(exposure)
	
	color.X = 1.0 - stdmath.Exp(-color.X)
	color.Y = 1.0 - stdmath.Exp(-color.Y)
	color.Z = 1.0 - stdmath.Exp(-color.Z)
	
	color.X = stdmath.Pow(color.X, 1.0/gamma)
	color.Y = stdmath.Pow(color.Y, 1.0/gamma)
	color.Z = stdmath.Pow(color.Z, 1.0/gamma)
	
	color.X = stdmath.Max(0.0, stdmath.Min(1.0, color.X))
	color.Y = stdmath.Max(0.0, stdmath.Min(1.0, color.Y))
	color.Z = stdmath.Max(0.0, stdmath.Min(1.0, color.Z))
	
	return color
}

func (r *ParallelRenderer) skyColor(ray geometry.Ray) math.Vec3 {
	return math.Vec3{X: 0.1, Y: 0.1, Z: 0.1}
}

func (r *ParallelRenderer) setupCamera(camera scene.Camera, width, height int) *scene.Camera {
	return &camera
}

func (r *ParallelRenderer) getRay(u, v float64, camera *scene.Camera) geometry.Ray {
	viewportHeight := 2.0
	viewportWidth := viewportHeight * float64(camera.AspectRatio)
	focalLength := 1.0
	
	origin := camera.Position
	horizontal := math.Vec3{X: viewportWidth, Y: 0, Z: 0}
	vertical := math.Vec3{X: 0, Y: viewportHeight, Z: 0}
	lowerLeftCorner := origin.Sub(horizontal.DivScalar(2)).Sub(vertical.DivScalar(2)).Sub(math.Vec3{X: 0, Y: 0, Z: focalLength})
	
	direction := lowerLeftCorner.Add(horizontal.MulScalar(u)).Add(vertical.MulScalar(v)).Sub(origin)
	
	return geometry.NewRay(origin, direction)
}

type RenderTask struct {
	startX, startY, endX, endY int
	width, height               int
	camera                      *scene.Camera
}

func (r *ParallelRenderer) createRenderTasks(width, height int, scene *scene.Scene, camera *scene.Camera) chan RenderTask {
	tasks := make(chan RenderTask, r.numWorkers*4)
	
	tileSize := 32
	numTilesX := (width + tileSize - 1) / tileSize
	numTilesY := (height + tileSize - 1) / tileSize
	
	go func() {
		for y := 0; y < numTilesY; y++ {
			for x := 0; x < numTilesX; x++ {
				startX := x * tileSize
				startY := y * tileSize
				endX := startX + tileSize
				if endX > width {
					endX = width
				}
				endY := startY + tileSize
				if endY > height {
					endY = height
				}
				
				task := RenderTask{
					startX:  startX,
					startY:  startY,
					endX:    endX,
					endY:    endY,
					width:   width,
					height:  height,
					camera:  camera,
				}
				
				tasks <- task
			}
		}
		close(tasks)
	}()
	
	return tasks
}

func (r *ParallelRenderer) SaveImage(img *image.RGBA, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, img)
}

func (r *ParallelRenderer) PrintASCIIPreview(img *image.RGBA) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	
	asciiChars := " .:-=+*#%@"
	
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x += 1 {
			r, g, b, _ := img.At(x, y).RGBA()
			brightness := float64(r+g+b) / 3.0
			charIndex := int(brightness * float64(len(asciiChars)-1) / 65535.0)
			if charIndex >= len(asciiChars) {
				charIndex = len(asciiChars) - 1
			}
			fmt.Printf("%c", asciiChars[charIndex])
		}
		fmt.Println()
	}
}

func (r *ParallelRenderer) SaveBenchmarkData(outputPath string) error {
	data, err := json.MarshalIndent(r.benchmarkData, "", "  ")
	if err != nil {
		return err
	}
	
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(outputPath, data, 0644)
} 