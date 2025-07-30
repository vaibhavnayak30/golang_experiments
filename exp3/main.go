// Create a main package
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Job is single unit of work
type Job struct {
	ID int
}

// Worker pool manages workers and job distribution 
type WorkerPool struct {
	numWorkers int
	jobs chan Job
	ctx context.Context
	cancel context.CancelFunc
	wg sync.WaitGroup
}

// NewWorkerPool creates new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs: make(chan Job, 10), //buffered channel
		ctx: ctx,
		cancel: cancel,
	}
}

// Worker handles jobs until context is cancelled 
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started/n", id)

	for {
		select {
		case <-wp.ctx.Done():
			fmt.Printf("Worker %d exiting...\n", id)
			return
		case job := <-wp.jobs:
			fmt.Printf("Worker %d processing job %d\n", id, job.ID)
			time.Sleep(time.Millisecond * 500)
			fmt.Printf("Worker %d finished job %d\n", id, job.ID)
		}
	}
}

// Start launches all worker goroutines 
func (wp *WorkerPool) Start() {
	for i:=1; i <= wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop gracefully shuts down all workers goroutines 
func (wp *WorkerPool) Stop() {
	fmt.Println("Stopping worker pool")
	// pass cancel signal
	wp.cancel() 
	// Wait for all workers to finish
	wp.wg.Wait()
	// Close channels to release resources
	close(wp.jobs)
	fmt.Println("All workers and channels closed and resources released")
}

// Submit a job to the job queue
func (wp *WorkerPool) Submit(job Job) {
	select {
	case <-wp.ctx.Done():
		fmt.Printf("WorkerPool closed. Dropping job %d\n", job.ID)
	default:
		wp.jobs <-job
	}
}

func main() {
	// Create a context that cancels on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wp := NewWorkerPool(3)
	// assign signal aware context
	wp.ctx = ctx
	wp.Start()

	// Job producer
	go func ()  {
		jobID := 1
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Job producer stopping")
				return
			default:
				wp.Submit(Job{ID: jobID})
				jobID++
				time.Sleep(300 * time.Millisecond)
		}
	}
		
	}()

	// Wait for signal cancellation 
	<-ctx.Done()
	fmt.Println("\nReceived signal cancellation")

	wp.Stop()
	fmt.Println("Graceful shutdown complete.")
}

