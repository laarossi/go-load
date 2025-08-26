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

	fmt.Printf("Pre-warming %d workers...\n", p.numWorkers)

	// Pre-warm and start all workers
	for i := 0; i < p.numWorkers; i++ {
		workerQuit := make(chan struct{})
		p.workers[i] = workerQuit

		go func(id int, quit chan struct{}) {
			fmt.Printf("Worker %d started and listening...\n", id)

			for {
				select {
				case task, ok := <-p.tasks:
					if !ok {
						// Tasks channel closed, finish up
						fmt.Printf("Worker %d: tasks channel closed, exiting\n", id)
						return
					}

					// Worker is now busy
					fmt.Printf("Worker %d: processing task\n", id)
					p.handler(task)
					p.wg.Done()
					fmt.Printf("Worker %d: finished task, back to listening\n", id)

				case <-quit:
					// Received shutdown signal
					fmt.Printf("Worker %d: received shutdown signal, exiting\n", id)
					return

				case <-p.quit:
					// Global shutdown signal
					fmt.Printf("Worker %d: global shutdown, exiting\n", id)
					return
				}
			}
		}(i, workerQuit)
	}

	p.started = true
	fmt.Printf("All %d workers pre-warmed and ready!\n", p.numWorkers)
}

// AddTask enqueues a task to the worker pool
func (p *WorkerPool[T]) AddTask(task T) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		fmt.Println("Worker pool not started yet")
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
		fmt.Println("Task added to queue")
		return true
	case <-time.After(time.Second):
		p.wg.Done() // Remove the count we added
		fmt.Println("Failed to add task: queue full or shutting down")
		return false
	}
}

// Stop gracefully shuts down the worker pool
// 1. Stops accepting new tasks
// 2. Waits for all busy workers to finish their current tasks
// 3. Terminates all workers
func (p *WorkerPool[T]) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		fmt.Println("Worker pool not started, nothing to stop")
		return
	}

	fmt.Println("Initiating graceful shutdown...")

	// Step 1: Signal that we're shutting down (prevents new tasks)
	close(p.quit)

	// Step 2: Close tasks channel so no new tasks can be added
	// Workers will finish processing current tasks
	close(p.tasks)

	fmt.Println("Waiting for all busy workers to finish their current tasks...")

	// Step 3: Wait for all busy workers to complete their work
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	// Wait with timeout to avoid hanging forever
	select {
	case <-done:
		fmt.Println("All busy workers finished their tasks")
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout waiting for workers to finish, forcing shutdown")
	}

	// Step 4: Send shutdown signal to any remaining workers
	fmt.Println("Terminating all workers...")
	for i, workerQuit := range p.workers {
		select {
		case <-workerQuit:
			// Already closed
		default:
			close(workerQuit)
			fmt.Printf("Sent termination signal to worker %d\n", i)
		}
	}

	// Give workers a moment to clean up
	time.Sleep(100 * time.Millisecond)

	p.started = false
	fmt.Println("Worker pool stopped successfully")
}

// GetStats returns current pool statistics
func (p *WorkerPool[T]) GetStats() (int, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	queueSize := len(p.tasks)
	return p.numWorkers, queueSize
}

// IsRunning returns whether the pool is currently running
func (p *WorkerPool[T]) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}
