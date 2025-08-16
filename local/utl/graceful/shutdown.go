package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ShutdownManager struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers []func() error
	mutex    sync.Mutex
}

var globalManager *ShutdownManager
var once sync.Once

func GetManager() *ShutdownManager {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		globalManager = &ShutdownManager{
			ctx:      ctx,
			cancel:   cancel,
			handlers: make([]func() error, 0),
		}
		
		go globalManager.watchSignals()
	})
	return globalManager
}

func (sm *ShutdownManager) watchSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	sm.Shutdown(30 * time.Second)
}

func (sm *ShutdownManager) AddHandler(handler func() error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.handlers = append(sm.handlers, handler)
}

func (sm *ShutdownManager) Context() context.Context {
	return sm.ctx
}

func (sm *ShutdownManager) AddGoroutine() {
	sm.wg.Add(1)
}

func (sm *ShutdownManager) GoroutineDone() {
	sm.wg.Done()
}

func (sm *ShutdownManager) RunGoroutine(fn func(ctx context.Context)) {
	sm.wg.Add(1)
	go func() {
		defer sm.wg.Done()
		fn(sm.ctx)
	}()
}

func (sm *ShutdownManager) Shutdown(timeout time.Duration) {
	sm.cancel()
	
	done := make(chan struct{})
	go func() {
		sm.wg.Wait()
		
		sm.mutex.Lock()
		for _, handler := range sm.handlers {
			handler()
		}
		sm.mutex.Unlock()
		
		close(done)
	}()
	
	select {
	case <-done:
	case <-time.After(timeout):
	}
}