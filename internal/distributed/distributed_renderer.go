package distributed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type DistributedRenderer struct {
	nodes        []string
	client       *http.Client
	ctx          context.Context
	cancel       context.CancelFunc
	
	nodeLoads    map[string]int
	loadMutex    sync.RWMutex
	
	remoteJobs   int64
	localJobs    int64
	failedJobs   int64
	startTime    time.Time
}

type RenderChunk struct {
	ID       int    `json:"id"`
	StartX   int    `json:"start_x"`
	EndX     int    `json:"end_x"`
	StartY   int    `json:"start_y"`
	EndY     int    `json:"end_y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Scene    string `json:"scene"`
	Priority int    `json:"priority"`
}

type RemoteResult struct {
	ChunkID  int       `json:"chunk_id"`
	Pixels   []Pixel   `json:"pixels"`
	Duration float64   `json:"duration"`
	Error    string    `json:"error,omitempty"`
	NodeID   string    `json:"node_id"`
}

type Pixel struct {
	X, Y int    `json:"x, y"`
	R, G, B, A uint8 `json:"r, g, b, a"`
}

type NodeInfo struct {
	ID           string  `json:"id"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  int64   `json:"memory_usage"`
	ActiveJobs   int     `json:"active_jobs"`
	MaxJobs      int     `json:"max_jobs"`
	LoadAverage  float64 `json:"load_average"`
}

func NewDistributedRenderer(ctx context.Context, nodes []string) *DistributedRenderer {
	ctx, cancel := context.WithCancel(ctx)
	
	return &DistributedRenderer{
		nodes:      nodes,
		client:     &http.Client{Timeout: 30 * time.Second},
		ctx:        ctx,
		cancel:     cancel,
		nodeLoads:  make(map[string]int),
		startTime:  time.Now(),
	}
}

func (dr *DistributedRenderer) RenderChunkRemotely(chunk RenderChunk, nodeAddr string) (*RemoteResult, error) {
	chunkData, err := json.Marshal(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chunk: %w", err)
	}
	
	req, err := http.NewRequestWithContext(dr.ctx, "POST", 
		fmt.Sprintf("http://%s/render", nodeAddr), 
		bytes.NewReader(chunkData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := dr.client.Do(req)
	if err != nil {
		atomic.AddInt64(&dr.failedJobs, 1)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	var result RemoteResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		atomic.AddInt64(&dr.failedJobs, 1)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	atomic.AddInt64(&dr.remoteJobs, 1)
	return &result, nil
}

func (dr *DistributedRenderer) GetOptimalNode() string {
	dr.loadMutex.RLock()
	defer dr.loadMutex.RUnlock()
	
	var bestNode string
	minLoad := int(^uint(0) >> 1) // Max int
	
	for node, load := range dr.nodeLoads {
		if load < minLoad {
			minLoad = load
			bestNode = node
		}
	}
	
	return bestNode
}

func (dr *DistributedRenderer) UpdateNodeLoad(nodeID string, load int) {
	dr.loadMutex.Lock()
	defer dr.loadMutex.Unlock()
	
	dr.nodeLoads[nodeID] = load
}

func (dr *DistributedRenderer) GetNodeInfo(nodeAddr string) (*NodeInfo, error) {
	req, err := http.NewRequestWithContext(dr.ctx, "GET", 
		fmt.Sprintf("http://%s/status", nodeAddr), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}
	
	resp, err := dr.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}
	defer resp.Body.Close()
	
	var nodeInfo NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodeInfo); err != nil {
		return nil, fmt.Errorf("failed to decode node info: %w", err)
	}
	
	return &nodeInfo, nil
}

func (dr *DistributedRenderer) DistributeWork(chunks []RenderChunk) ([]RemoteResult, error) {
	var results []RemoteResult
	var wg sync.WaitGroup
	resultChan := make(chan RemoteResult, len(chunks))
	errorChan := make(chan error, len(chunks))
	
	for i, chunk := range chunks {
		wg.Add(1)
		go func(chunk RenderChunk, index int) {
			defer wg.Done()
			
			nodeAddr := dr.GetOptimalNode()
			if nodeAddr == "" {
				errorChan <- fmt.Errorf("no available nodes")
				return
			}
			
			result, err := dr.RenderChunkRemotely(chunk, nodeAddr)
			if err != nil {
				errorChan <- fmt.Errorf("failed to render chunk %d: %w", chunk.ID, err)
				return
			}
			
			resultChan <- *result
		}(chunk, i)
	}
	
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()
	
	for result := range resultChan {
		results = append(results, result)
	}
	
	select {
	case err := <-errorChan:
		return results, err
	default:
		return results, nil
	}
}

func (dr *DistributedRenderer) GetStats() map[string]interface{} {
	elapsed := time.Since(dr.startTime)
	
	return map[string]interface{}{
		"remote_jobs":    atomic.LoadInt64(&dr.remoteJobs),
		"local_jobs":     atomic.LoadInt64(&dr.localJobs),
		"failed_jobs":    atomic.LoadInt64(&dr.failedJobs),
		"total_nodes":    len(dr.nodes),
		"elapsed_time":   elapsed,
		"success_rate":   calculateSuccessRate(dr.remoteJobs, dr.failedJobs),
	}
}

func calculateSuccessRate(remote, failed int64) float64 {
	total := remote + failed
	if total == 0 {
		return 100.0
	}
	return float64(remote) / float64(total) * 100
}

type RemoteRenderServer struct {
	port     string
	renderer interface{} // Local renderer instance
	server   *http.Server
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewRemoteRenderServer(port string, renderer interface{}) *RemoteRenderServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RemoteRenderServer{
		port:     port,
		renderer: renderer,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (rrs *RemoteRenderServer) Start() error {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/render", rrs.handleRender)
	
	mux.HandleFunc("/status", rrs.handleStatus)
	
	rrs.server = &http.Server{
		Addr:    ":" + rrs.port,
		Handler: mux,
	}
	
	return rrs.server.ListenAndServe()
}

func (rrs *RemoteRenderServer) Stop() error {
	rrs.cancel()
	return rrs.server.Shutdown(context.Background())
}

func (rrs *RemoteRenderServer) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var chunk RenderChunk
	if err := json.NewDecoder(r.Body).Decode(&chunk); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	start := time.Now()
	
	time.Sleep(100 * time.Millisecond)
	
	result := RemoteResult{
		ChunkID:  chunk.ID,
		Duration: time.Since(start).Seconds(),
		NodeID:   "node-" + rrs.port,
		Pixels:   []Pixel{}, // In real implementation, this would contain actual pixels
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (rrs *RemoteRenderServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	nodeInfo := NodeInfo{
		ID:          "node-" + rrs.port,
		CPUUsage:    50.0, // In real implementation, get actual CPU usage
		MemoryUsage: 1024 * 1024 * 100, // 100MB
		ActiveJobs:  5,
		MaxJobs:     10,
		LoadAverage: 0.5,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeInfo)
}

type LoadBalancer struct {
	nodes    []string
	strategy LoadBalancingStrategy
	mu       sync.RWMutex
}

type LoadBalancingStrategy interface {
	SelectNode(nodes []string, loads map[string]int) string
}

type RoundRobinStrategy struct {
	current int
	mu      sync.Mutex
}

func (rr *RoundRobinStrategy) SelectNode(nodes []string, loads map[string]int) string {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	
	if len(nodes) == 0 {
		return ""
	}
	
	node := nodes[rr.current]
	rr.current = (rr.current + 1) % len(nodes)
	return node
}

type LeastConnectionsStrategy struct{}

func (lc *LeastConnectionsStrategy) SelectNode(nodes []string, loads map[string]int) string {
	if len(nodes) == 0 {
		return ""
	}
	
	var bestNode string
	minLoad := int(^uint(0) >> 1)
	
	for _, node := range nodes {
		if load, exists := loads[node]; exists && load < minLoad {
			minLoad = load
			bestNode = node
		}
	}
	
	return bestNode
}

func NewLoadBalancer(nodes []string, strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		nodes:    nodes,
		strategy: strategy,
	}
}

func (lb *LoadBalancer) GetNode() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	loads := make(map[string]int)
	for _, node := range lb.nodes {
		loads[node] = 0 // Placeholder
	}
	
	return lb.strategy.SelectNode(lb.nodes, loads)
} 