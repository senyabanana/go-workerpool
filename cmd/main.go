package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/senyabanana/go-workerpool/internal/fileinput"
	"github.com/senyabanana/go-workerpool/internal/workerpool"
)

const (
	shadesOfGoExample = "50_shades_of_go_example.txt"
)

func main() {
	pool := workerpool.NewWorkerPool()

	tasks, err := fileinput.LoadFromFile(shadesOfGoExample)
	if err != nil {
		fmt.Printf("failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	go startTaskFeeder(pool, tasks, stopChan, doneChan)
	go startRandomScaling(pool, stopChan)

	select {
	case <-time.After(50 * time.Second):
		fmt.Println("Waiting time is up")
	case <-doneChan:
		fmt.Println("All tasks from file have been sent")
	}

	close(stopChan)
	pool.Shutdown()
}

func startTaskFeeder(pool *workerpool.Pool, tasks []string, stopChan, doneChan chan struct{}) {
	defer close(doneChan)

	for _, task := range tasks {
		select {
		case <-stopChan:
			return
		case pool.TaskChan() <- task:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func startRandomScaling(pool *workerpool.Pool, stopChan chan struct{}) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			action := rand.Intn(2)
			switch action {
			case 0:
				pool.AddWorker()
			case 1:
				pool.MU.Lock()
				poolLen := pool.LenWorkers()
				if poolLen > 0 {
					for id := range poolLen {
						pool.MU.Unlock()
						pool.RemoveWorker(id)
						break
					}
				} else {
					pool.MU.Unlock()
				}
			}
		}
	}
}
