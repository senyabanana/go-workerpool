package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	shadesOfGoExample = "50_shades_of_go_example.txt"
)

type Worker struct {
	ID     int
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewWorker(id int, cancel context.CancelFunc, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:     id,
		cancel: cancel,
		wg:     wg,
	}
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

	worker := NewWorker(id, cancel, &p.wg)

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

func loadFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return lines, nil
}

func main() {
	pool := NewWorkerPool()

	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()

	tasks, err := loadFromFile(shadesOfGoExample)
	if err != nil {
		fmt.Printf("failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		for _, task := range tasks {
			select {
			case <-stopChan:
				return
			case pool.taskChan <- task:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	select {
	case <-time.After(50 * time.Second):
		fmt.Println("Waiting time is up")
	case <-doneChan:
		fmt.Println("All jobs from the file have been sent")
	}

	time.Sleep(5 * time.Second)
	pool.RemoveWorker(1)

	time.Sleep(15 * time.Second)
	pool.RemoveWorker(3)

	time.Sleep(25 * time.Second)

	close(stopChan)
	pool.Shutdown()
}
