package workerpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/senyabanana/go-workerpool/internal/worker"
)

type Pool struct {
	taskChan chan string
	workers  map[int]*worker.Worker
	MU       sync.Mutex
	wg       sync.WaitGroup
	nextID   int
}

func NewWorkerPool() *Pool {
	return &Pool{
		taskChan: make(chan string),
		workers:  make(map[int]*worker.Worker),
	}
}

func (p *Pool) AddWorker() {
	p.MU.Lock()
	defer p.MU.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w := worker.NewWorker(p.nextID, cancel, &p.wg)
	p.workers[p.nextID] = w
	p.nextID++

	w.Start(ctx, p.taskChan)
	fmt.Printf("Worker #%d added\n", w.ID)
}

func (p *Pool) RemoveWorker(id int) {
	p.MU.Lock()
	defer p.MU.Unlock()

	w, ok := p.workers[id]
	if !ok {
		fmt.Printf("Worker #%d not found\n", id)
		return
	}

	w.Cancel()
	delete(p.workers, id)
	fmt.Printf("Worker #%d removed\n", id)
}

func (p *Pool) Shutdown() {
	p.MU.Lock()
	defer p.MU.Unlock()

	for id, w := range p.workers {
		fmt.Printf("Worker #%d shutting down\n", id)
		w.Cancel()
	}

	p.workers = make(map[int]*worker.Worker)
	//close(p.taskChan)

	p.wg.Wait()
	fmt.Println("All workers have shut down")
}

func (p *Pool) TaskChan() chan string {
	return p.taskChan
}

func (p *Pool) LenWorkers() int {
	return len(p.workers)
}
