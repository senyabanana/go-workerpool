package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	ID     int
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (w *Worker) Start(ctx context.Context, taskChan <-chan string) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		fmt.Printf("Worker %d starting\n", w.ID)
		for {
			select {
			case task, ok := <-taskChan:
				if !ok {
					fmt.Printf("Worker #%d: the task channel is closed\n", w.ID)
					return
				}
				fmt.Printf("Worker #%d: %s\n", w.ID, task)
				time.Sleep(1 * time.Second)
			case <-ctx.Done():
				fmt.Printf("Worker #%d: context is done\n", w.ID)
				return
			}
		}
	}()
}

type WorkerPool struct {
	taskChan chan string
	workers  map[int]*Worker
	mu       sync.Mutex
	wg       sync.WaitGroup
	nextID   int
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		taskChan: make(chan string),
		workers:  make(map[int]*Worker),
	}
}

func (p *WorkerPool) AddWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.nextID
	p.nextID++

	ctx, cancel := context.WithCancel(context.Background())

	worker := &Worker{
		ID:     id,
		cancel: cancel,
		wg:     &p.wg,
	}

	p.workers[id] = worker
	worker.Start(ctx, p.taskChan)

	fmt.Printf("Worker #%d added\n", id)
}

func (p *WorkerPool) RemoveWorker(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	worker, ok := p.workers[id]
	if !ok {
		fmt.Printf("Worker #%d not found\n", id)
		return
	}

	worker.cancel()
	delete(p.workers, id)

	fmt.Printf("Worker #%d removed\n", id)
}

func (p *WorkerPool) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, worker := range p.workers {
		fmt.Printf("Worker #%d shutting down\n", id)
		worker.cancel()
	}

	p.workers = make(map[int]*Worker)
	//close(p.taskChan)

	p.wg.Wait()
	fmt.Println("All workers have shut down")
}

func main() {
	pool := NewWorkerPool()

	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()

	stopChan := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			select {
			case <-stopChan:
				return
			case pool.taskChan <- fmt.Sprintf("Task #%d", i):
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second)
	pool.RemoveWorker(0)

	time.Sleep(2 * time.Second)
	pool.RemoveWorker(2)

	time.Sleep(3 * time.Second)

	close(stopChan)
	pool.Shutdown()
}
