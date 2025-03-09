package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type GracefulShutdown struct {
	ctx           context.Context
	cancel        context.CancelFunc
	shutdownChan  chan os.Signal
	cleanupFuncs  []CleanupFunc
	mu            sync.Mutex
	
	isShuttingDown bool
	shutdownWg     sync.WaitGroup
	
	shutdownTimeout time.Duration
	cleanupTimeout  time.Duration
}

type CleanupFunc func(ctx context.Context) error

type ShutdownHook struct {
	Name     string
	Priority int
	Func     CleanupFunc
}

func NewGracefulShutdown(ctx context.Context) *GracefulShutdown {
	ctx, cancel := context.WithCancel(ctx)
	
	return &GracefulShutdown{
		ctx:             ctx,
		cancel:          cancel,
		shutdownChan:    make(chan os.Signal, 1),
		cleanupFuncs:    make([]CleanupFunc, 0),
		shutdownTimeout: 30 * time.Second,
		cleanupTimeout:  10 * time.Second,
	}
}

func (gs *GracefulShutdown) Start() {
	signal.Notify(gs.shutdownChan, os.Interrupt, syscall.SIGTERM)
	
	go gs.handleShutdown()
}

func (gs *GracefulShutdown) handleShutdown() {
	select {
	case sig := <-gs.shutdownChan:
		fmt.Printf("Received signal %v, initiating graceful shutdown...\n", sig)
		gs.Shutdown()
	case <-gs.ctx.Done():
		fmt.Println("Context cancelled, initiating shutdown...")
		gs.Shutdown()
	}
}

func (gs *GracefulShutdown) Shutdown() {
	gs.mu.Lock()
	if gs.isShuttingDown {
		gs.mu.Unlock()
		return
	}
	gs.isShuttingDown = true
	gs.mu.Unlock()
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
	defer cancel()
	
	fmt.Println("Starting graceful shutdown...")
	
	gs.cancel()
	
	done := make(chan struct{})
	go func() {
		gs.shutdownWg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		fmt.Println("Graceful shutdown completed successfully")
	case <-shutdownCtx.Done():
		fmt.Println("Shutdown timeout reached, forcing exit")
		os.Exit(1)
	}
}

func (gs *GracefulShutdown) AddCleanupFunc(name string, priority int, cleanupFunc CleanupFunc) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	gs.shutdownWg.Add(1)
	
	go func() {
		defer gs.shutdownWg.Done()
		
		<-gs.ctx.Done()
		
		cleanupCtx, cancel := context.WithTimeout(context.Background(), gs.cleanupTimeout)
		defer cancel()
		
		fmt.Printf("Executing cleanup: %s (priority: %d)\n", name, priority)
		
		if err := cleanupFunc(cleanupCtx); err != nil {
			fmt.Printf("Error during cleanup %s: %v\n", name, err)
		} else {
			fmt.Printf("Cleanup completed: %s\n", name)
		}
	}()
}

func (gs *GracefulShutdown) IsShuttingDown() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.isShuttingDown
}

func (gs *GracefulShutdown) GetContext() context.Context {
	return gs.ctx
}

type ResourceManager struct {
	resources map[string]Resource
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type Resource interface {
	Close() error
	Name() string
}

type ManagedResource struct {
	resource Resource
	manager  *ResourceManager
}

func NewResourceManager(ctx context.Context) *ResourceManager {
	ctx, cancel := context.WithCancel(ctx)
	
	return &ResourceManager{
		resources: make(map[string]Resource),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (rm *ResourceManager) AddResource(name string, resource Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.resources[name] = resource
}

func (rm *ResourceManager) RemoveResource(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if resource, exists := rm.resources[name]; exists {
		resource.Close()
		delete(rm.resources, name)
	}
}

func (rm *ResourceManager) GetResource(name string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	resource, exists := rm.resources[name]
	return resource, exists
}

func (rm *ResourceManager) CloseAll() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	var errors []error
	
	for name, resource := range rm.resources {
		if err := resource.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", name, err))
		}
	}
	
	rm.resources = make(map[string]Resource)
	
	if len(errors) > 0 {
		return fmt.Errorf("errors during resource cleanup: %v", errors)
	}
	
	return nil
}

func (rm *ResourceManager) GetResourceCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	return len(rm.resources)
}

type ContextShutdown struct {
	ctx           context.Context
	cancel        context.CancelFunc
	shutdownFuncs []ShutdownFunc
	mu            sync.Mutex
	
	shutdownComplete chan struct{}
}

type ShutdownFunc func(ctx context.Context) error

func NewContextShutdown(ctx context.Context) *ContextShutdown {
	ctx, cancel := context.WithCancel(ctx)
	
	return &ContextShutdown{
		ctx:              ctx,
		cancel:           cancel,
		shutdownFuncs:    make([]ShutdownFunc, 0),
		shutdownComplete: make(chan struct{}),
	}
}

func (cs *ContextShutdown) AddShutdownFunc(fn ShutdownFunc) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.shutdownFuncs = append(cs.shutdownFuncs, fn)
}

func (cs *ContextShutdown) Shutdown(timeout time.Duration) error {
	cs.mu.Lock()
	if cs.shutdownComplete == nil {
		cs.mu.Unlock()
		return fmt.Errorf("shutdown already completed")
	}
	close(cs.shutdownComplete)
	cs.mu.Unlock()
	
	cs.cancel()
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	var wg sync.WaitGroup
	errorChan := make(chan error, len(cs.shutdownFuncs))
	
	for _, fn := range cs.shutdownFuncs {
		wg.Add(1)
		go func(shutdownFunc ShutdownFunc) {
			defer wg.Done()
			
			if err := shutdownFunc(shutdownCtx); err != nil {
				select {
				case errorChan <- err:
				default:
				}
			}
		}(fn)
	}
	
	wg.Wait()
	close(errorChan)
	
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	return nil
}

func (cs *ContextShutdown) IsShutdownComplete() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	select {
	case <-cs.shutdownComplete:
		return true
	default:
		return false
	}
}

type SignalHandler struct {
	signals map[os.Signal]SignalAction
	ctx     context.Context
	cancel  context.CancelFunc
}

type SignalAction func(sig os.Signal) error

func NewSignalHandler(ctx context.Context) *SignalHandler {
	ctx, cancel := context.WithCancel(ctx)
	
	return &SignalHandler{
		signals: make(map[os.Signal]SignalAction),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (sh *SignalHandler) RegisterSignal(sig os.Signal, action SignalAction) {
	sh.signals[sig] = action
}

func (sh *SignalHandler) Start() {
	sigChan := make(chan os.Signal, 1)
	
	for sig := range sh.signals {
		signal.Notify(sigChan, sig)
	}
	
	go func() {
		for {
			select {
			case sig := <-sigChan:
				if action, exists := sh.signals[sig]; exists {
					if err := action(sig); err != nil {
						fmt.Printf("Error handling signal %v: %v\n", sig, err)
					}
				}
			case <-sh.ctx.Done():
				return
			}
		}
	}()
}

func (sh *SignalHandler) Stop() {
	sh.cancel()
} 