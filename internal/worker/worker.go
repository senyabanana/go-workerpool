package worker

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

func (w *Worker) Cancel() {
	w.cancel()
}
