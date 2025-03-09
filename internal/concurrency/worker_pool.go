package concurrency

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type RenderJob struct {
	ID       int
	StartX   int
	EndX     int
	StartY   int
	EndY     int
	Width    int
	Height   int
	Priority int
	Scene    interface{}
	Camera   interface{}
}

type RenderResult struct {
	JobID    int
	Pixels   []Pixel
	StartX   int
	StartY   int
	Duration time.Duration
	Error    error
}

type Pixel struct {
	X, Y int
	R, G, B, A uint8
}

type WorkerPool struct {
	workers       int
	jobQueue      chan RenderJob
	resultQueue   chan RenderResult
	workerWg      sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	
	activeWorkers int32
	completedJobs int64
	totalJobs     int64
	startTime     time.Time
	
	workStealers []*WorkStealer
	globalQueue  chan RenderJob
	
	rayPool      *sync.Pool
	hitPool      *sync.Pool
	vectorPool   *sync.Pool
}

type WorkStealer struct {
	localQueue []RenderJob
	globalQueue chan RenderJob
	mu         sync.Mutex
	workerID   int
}

func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		workers:     workers,
		jobQueue:    make(chan RenderJob, workers*2),
		resultQueue: make(chan RenderResult, workers*2),
		ctx:         ctx,
		cancel:      cancel,
		startTime:   time.Now(),
		globalQueue: make(chan RenderJob, workers*4),
		workStealers: make([]*WorkStealer, workers),
	}
	
	pool.rayPool = &sync.Pool{
		New: func() interface{} {
			return &Ray{}
		},
	}
	
	pool.hitPool = &sync.Pool{
		New: func() interface{} {
			return &HitRecord{}
		},
	}
	
	pool.vectorPool = &sync.Pool{
		New: func() interface{} {
			return &Vec3{}
		},
	}
	
	for i := 0; i < workers; i++ {
		pool.workStealers[i] = &WorkStealer{
			localQueue:  make([]RenderJob, 0, 10),
			globalQueue: pool.globalQueue,
			workerID:    i,
		}
	}
	
	return pool
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.workerWg.Add(1)
		go wp.worker(i)
	}
	
	go wp.collectMetrics()
}

func (wp *WorkerPool) worker(id int) {
	defer wp.workerWg.Done()
	
	workStealer := wp.workStealers[id]
	
	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			wp.processJob(job, workStealer)
		default:
			if stolenJob := workStealer.StealWork(); stolenJob.ID != 0 {
				wp.processJob(stolenJob, workStealer)
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

func (wp *WorkerPool) processJob(job RenderJob, workStealer *WorkStealer) {
	atomic.AddInt32(&wp.activeWorkers, 1)
	defer atomic.AddInt32(&wp.activeWorkers, -1)
	
	start := time.Now()
	
	result := RenderResult{
		JobID:    job.ID,
		StartX:   job.StartX,
		StartY:   job.StartY,
		Duration: time.Since(start),
	}
	
	select {
	case wp.resultQueue <- result:
	case <-wp.ctx.Done():
		return
	}
	
	atomic.AddInt64(&wp.completedJobs, 1)
}

func (ws *WorkStealer) StealWork() RenderJob {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	if len(ws.localQueue) > 0 {
		job := ws.localQueue[len(ws.localQueue)-1]
		ws.localQueue = ws.localQueue[:len(ws.localQueue)-1]
		return job
	}
	
	select {
	case job := <-ws.globalQueue:
		return job
	default:
		return RenderJob{} // No work available
	}
}

func (wp *WorkerPool) SubmitJob(job RenderJob) error {
	select {
	case wp.jobQueue <- job:
		atomic.AddInt64(&wp.totalJobs, 1)
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	}
}

func (wp *WorkerPool) GetResult() (RenderResult, bool) {
	select {
	case result := <-wp.resultQueue:
		return result, true
	case <-wp.ctx.Done():
		return RenderResult{}, false
	}
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobQueue)
	wp.workerWg.Wait()
	close(wp.resultQueue)
}

func (wp *WorkerPool) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			wp.reportMetrics()
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) reportMetrics() {
	active := atomic.LoadInt32(&wp.activeWorkers)
	completed := atomic.LoadInt64(&wp.completedJobs)
	total := atomic.LoadInt64(&wp.totalJobs)
	elapsed := time.Since(wp.startTime)
	
	if elapsed > 0 {
		rate := float64(completed) / elapsed.Seconds()
		progress := float64(completed) / float64(total) * 100
		
		_ = rate
		_ = progress
		_ = active
	}
}

func (wp *WorkerPool) GetStats() map[string]interface{} {
	active := atomic.LoadInt32(&wp.activeWorkers)
	completed := atomic.LoadInt64(&wp.completedJobs)
	total := atomic.LoadInt64(&wp.totalJobs)
	elapsed := time.Since(wp.startTime)
	
	rate := float64(0)
	if elapsed > 0 {
		rate = float64(completed) / elapsed.Seconds()
	}
	
	progress := float64(0)
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}
	
	return map[string]interface{}{
		"active_workers": active,
		"total_workers":  wp.workers,
		"completed_jobs": completed,
		"total_jobs":     total,
		"progress":       progress,
		"rate_per_sec":   rate,
		"elapsed_time":   elapsed,
	}
}

func (wp *WorkerPool) GetRay() *Ray {
	return wp.rayPool.Get().(*Ray)
}

func (wp *WorkerPool) PutRay(ray *Ray) {
	wp.rayPool.Put(ray)
}

func (wp *WorkerPool) GetHitRecord() *HitRecord {
	return wp.hitPool.Get().(*HitRecord)
}

func (wp *WorkerPool) PutHitRecord(hit *HitRecord) {
	wp.hitPool.Put(hit)
}

func (wp *WorkerPool) GetVec3() *Vec3 {
	return wp.vectorPool.Get().(*Vec3)
}

func (wp *WorkerPool) PutVec3(vec *Vec3) {
	wp.vectorPool.Put(vec)
}

type Ray struct{}
type HitRecord struct{}
type Vec3 struct{} 