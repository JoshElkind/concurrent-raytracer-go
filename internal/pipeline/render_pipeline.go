package pipeline

import (
	"context"
	"sync"
	"time"
)

type Ray struct {
	Origin    Vec3
	Direction Vec3
}

type Vec3 struct {
	X, Y, Z float64
}

type Intersection struct {
	Ray       Ray
	Point     Vec3
	Normal    Vec3
	T         float64
	Material  interface{}
	JobID     int
}

type ShadedPixel struct {
	X, Y int
	R, G, B, A uint8
	JobID int
}

type RenderPipeline struct {
	ctx context.Context
	
	rayGen       chan Ray
	intersections chan Intersection
	shadedPixels chan ShadedPixel
	finalImage   chan []ShadedPixel
	
	rayGenWorkers     int
	intersectionWorkers int
	shadingWorkers    int
	
	wg sync.WaitGroup
	done chan struct{}
	
	raysGenerated    int64
	intersectionsFound int64
	pixelsShaded     int64
	startTime        time.Time
}

func NewRenderPipeline(ctx context.Context, rayWorkers, intersectionWorkers, shadingWorkers int) *RenderPipeline {
	if rayWorkers <= 0 {
		rayWorkers = 4
	}
	if intersectionWorkers <= 0 {
		intersectionWorkers = 8
	}
	if shadingWorkers <= 0 {
		shadingWorkers = 4
	}
	
	return &RenderPipeline{
		ctx:                 ctx,
		rayGen:             make(chan Ray, 1000),
		intersections:      make(chan Intersection, 1000),
		shadedPixels:       make(chan ShadedPixel, 1000),
		finalImage:         make(chan []ShadedPixel, 100),
		rayGenWorkers:      rayWorkers,
		intersectionWorkers: intersectionWorkers,
		shadingWorkers:     shadingWorkers,
		done:               make(chan struct{}),
		startTime:          time.Now(),
	}
}

func (rp *RenderPipeline) Start() {
	for i := 0; i < rp.rayGenWorkers; i++ {
		rp.wg.Add(1)
		go rp.rayGenerator(i)
	}
	
	for i := 0; i < rp.intersectionWorkers; i++ {
		rp.wg.Add(1)
		go rp.intersectionWorker(i)
	}
	
	for i := 0; i < rp.shadingWorkers; i++ {
		rp.wg.Add(1)
		go rp.shadingWorker(i)
	}
	
	rp.wg.Add(1)
	go rp.imageAssembler()
}

func (rp *RenderPipeline) rayGenerator(workerID int) {
	defer rp.wg.Done()
	
	for {
		select {
		case <-rp.ctx.Done():
			return
		case <-rp.done:
			return
		default:
			ray := Ray{
				Origin:    Vec3{X: 0, Y: 0, Z: -5},
				Direction: Vec3{X: 0, Y: 0, Z: 1},
			}
			
			select {
			case rp.rayGen <- ray:
			case <-rp.ctx.Done():
				return
			case <-rp.done:
				return
			}
			
			time.Sleep(1 * time.Microsecond)
		}
	}
}

func (rp *RenderPipeline) intersectionWorker(workerID int) {
	defer rp.wg.Done()
	
	for ray := range rp.rayGen {
		select {
		case <-rp.ctx.Done():
			return
		case <-rp.done:
			return
		default:
			intersection := Intersection{
				Ray:   ray,
				Point: Vec3{X: 0, Y: 0, Z: 0},
				Normal: Vec3{X: 0, Y: 0, Z: 1},
				T:     5.0,
			}
			
			select {
			case rp.intersections <- intersection:
			case <-rp.ctx.Done():
				return
			case <-rp.done:
				return
			}
			
			time.Sleep(10 * time.Microsecond)
		}
	}
}

func (rp *RenderPipeline) shadingWorker(workerID int) {
	defer rp.wg.Done()
	
	for _ = range rp.intersections {
		select {
		case <-rp.ctx.Done():
			return
		case <-rp.done:
			return
		default:
			pixel := ShadedPixel{
				X: 0, Y: 0,
				R: 255, G: 255, B: 255, A: 255,
			}
			
			select {
			case rp.shadedPixels <- pixel:
			case <-rp.ctx.Done():
				return
			case <-rp.done:
				return
			}
			
			time.Sleep(5 * time.Microsecond)
		}
	}
}

func (rp *RenderPipeline) imageAssembler() {
	defer rp.wg.Done()
	
	var pixels []ShadedPixel
	
	for pixel := range rp.shadedPixels {
		select {
		case <-rp.ctx.Done():
			return
		case <-rp.done:
			return
		default:
			pixels = append(pixels, pixel)
			
			if len(pixels) >= 1000 {
				select {
				case rp.finalImage <- pixels:
					pixels = pixels[:0] // Reset slice
				case <-rp.ctx.Done():
					return
				case <-rp.done:
					return
				}
			}
		}
	}
	
	if len(pixels) > 0 {
		select {
		case rp.finalImage <- pixels:
		case <-rp.ctx.Done():
			return
		case <-rp.done:
			return
		}
	}
}

func (rp *RenderPipeline) Stop() {
	close(rp.done)
	close(rp.rayGen)
	close(rp.intersections)
	close(rp.shadedPixels)
	rp.wg.Wait()
	close(rp.finalImage)
}

func (rp *RenderPipeline) GetFinalImage() <-chan []ShadedPixel {
	return rp.finalImage
}

func (rp *RenderPipeline) GetStats() map[string]interface{} {
	elapsed := time.Since(rp.startTime)
	
	return map[string]interface{}{
		"rays_generated":     rp.raysGenerated,
		"intersections_found": rp.intersectionsFound,
		"pixels_shaded":      rp.pixelsShaded,
		"elapsed_time":       elapsed,
		"ray_gen_workers":    rp.rayGenWorkers,
		"intersection_workers": rp.intersectionWorkers,
		"shading_workers":    rp.shadingWorkers,
	}
}

type AdaptivePipeline struct {
	*RenderPipeline
	metricsChan chan PipelineMetrics
	adjustmentTicker *time.Ticker
}

type PipelineMetrics struct {
	RayGenQueueLen     int
	IntersectionQueueLen int
	ShadingQueueLen    int
	CPUUsage           float64
	MemoryUsage        int64
}

func NewAdaptivePipeline(ctx context.Context) *AdaptivePipeline {
	pipeline := NewRenderPipeline(ctx, 4, 8, 4)
	
	adaptive := &AdaptivePipeline{
		RenderPipeline:     pipeline,
		metricsChan:        make(chan PipelineMetrics, 10),
		adjustmentTicker:   time.NewTicker(5 * time.Second),
	}
	
	return adaptive
}

func (ap *AdaptivePipeline) Start() {
	ap.RenderPipeline.Start()
	
	go ap.adaptiveAdjustment()
}

func (ap *AdaptivePipeline) adaptiveAdjustment() {
	for {
		select {
		case <-ap.adjustmentTicker.C:
			ap.adjustWorkerCounts()
		case <-ap.ctx.Done():
			return
		}
	}
}

func (ap *AdaptivePipeline) adjustWorkerCounts() {
	
	_ = ap.metricsChan
}

func (ap *AdaptivePipeline) Stop() {
	ap.adjustmentTicker.Stop()
	ap.RenderPipeline.Stop()
} 