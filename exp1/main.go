package main

import (
	"context"
	"fmt"
	"sync"
	"math/rand"
	"time"
)

// Worker function: Processes job from the job channel 
// Listens for job values or context cancellations to gracefully exit 
func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d exiting...\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				// If channel is closed or empty, exit 
				fmt.Printf("Worker %d job channel closed\n", id)
			}

			fmt.Printf("Worker %d started job %d\n", id, job)

			// Simulate processing time with delay 
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)+500))

			fmt.Printf("Worker %d: finished job %d\n", id, job)

		}
	}

}

// Sends new jobs into "jobs" channel until the context is cancelled 
func produceJobs(ctx context.Context, jobs chan<- int) {
	jobID := 1
	for {
		select {
		case <-ctx.Done():
			//Stop producing jobs, close the channel so workers know 
			fmt.Println("Job producer exiting...")
			close(jobs)
			return
		
		default:
			// Send a job into a channel 
			jobs <- jobID
			fmt.Printf("Produced job %d\n", jobID)
			jobID++
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func main() {
	// Create a context that cancels everything after 5 seconds 
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	// Ensures context is cancelled before main exits 
	defer cancel()

	// Buffered channel to hold pending jobs 
	jobs := make(chan int, 10)

	var wg sync.WaitGroup

	// Number of concurrent worker goroutines 
	numWorkers := 3

	// Launch multiple workers 
	for i:=1; i<=numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, &wg)
	}

	// Start the job producer 
	go produceJobs(ctx, jobs)

	// Wait for all wokers to finish 
	wg.Wait()

	fmt.Println("All workers are done")
}