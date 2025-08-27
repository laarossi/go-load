package utils

import (
	"fmt"
	"sync"
	"time"
)

type WorkerPool[T any] struct {
	tasks      chan T
	wg         sync.WaitGroup
	numWorkers int
	handler    func(T)
	quit       chan struct{}
	workers    []chan struct{} // individual quit channels for each worker
	started    bool
	mu         sync.RWMutex
}

func NewWorkerPool[T any](numWorkers int, handler func(T)) *WorkerPool[T] {
	return &WorkerPool[T]{
		tasks:      make(chan T, 100), // buffered queue for events
		numWorkers: numWorkers,
		handler:    handler,
		quit:       make(chan struct{}),
		workers:    make([]chan struct{}, numWorkers),
	}
}

// Start pre-warms the workers and starts them listening to the queue
func (p *WorkerPool[T]) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		fmt.Println("Worker pool already started")
		return
	}

	for i := 0; i < p.numWorkers; i++ {
		workerQuit := make(chan struct{})
		p.workers[i] = workerQuit

		go func(id int, quit chan struct{}) {
			for {
				select {
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					p.handler(task)
					p.wg.Done()

				case <-quit:
					return

				case <-p.quit:
					return
				}
			}
		}(i, workerQuit)
	}

	p.started = true
}

// AddTask enqueues a task to the worker pool
func (p *WorkerPool[T]) AddTask(task T) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return false
	}

	select {
	case <-p.quit:
		fmt.Println("Worker pool is shutting down, cannot add task")
		return false
	default:
	}

	p.wg.Add(1)

	// Non-blocking send to avoid deadlock during shutdown
	select {
	case p.tasks <- task:
		return true
	case <-time.After(time.Second):
		p.wg.Done()
		return false
	}
}

func (p *WorkerPool[T]) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}

	close(p.quit)
	close(p.tasks)
	fmt.Println("Waiting for all busy workers to finish their current tasks...")

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	time.After(30 * time.Second)
	for _, workerQuit := range p.workers {
		select {
		case <-workerQuit:
			// Already closed
		default:
			close(workerQuit)
		}
	}

	time.Sleep(100 * time.Millisecond)
	p.started = false
}

func (p *WorkerPool[T]) GetStats() (int, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	queueSize := len(p.tasks)
	return p.numWorkers, queueSize
}

func (p *WorkerPool[T]) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}
