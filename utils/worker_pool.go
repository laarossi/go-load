package utils

import (
	"fmt"
	"sync"
)

type WorkerPool[T any] struct {
	tasks      chan T
	wg         sync.WaitGroup
	numWorkers int
	handler    func(T) // how to process a task of type T
	quit       chan struct{}
}

func NewWorkerPool[T any](numWorkers int, handler func(T)) WorkerPool[T] {
	return WorkerPool[T]{
		tasks:      make(chan T, 100), // buffered queue
		numWorkers: numWorkers,
		handler:    handler,
		quit:       make(chan struct{}),
	}
}

func (p *WorkerPool[T]) Start() {
	for i := 0; i < p.numWorkers; i++ {
		go func(id int) {
			for {
				select {
				case task := <-p.tasks:
					p.handler(task)
					p.wg.Done()
				case <-p.quit:
					fmt.Printf("Worker %d stopping...\n", id)
					return
				}
			}
		}(i)
	}
}

// AddTask enqueues a task
func (p *WorkerPool[T]) AddTask(task T) {
	p.wg.Add(1)
	p.tasks <- task
}

func (p *WorkerPool[T]) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool[T]) Stop() {
	close(p.quit)
}
