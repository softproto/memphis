package contractor

import (
	"context"
	"log"
	"strconv"
	"sync"
)

const WorkersCount = 5

type WorkerPool struct {
	workersCount int
	results      chan Result
}

func Run(ctx context.Context, wg *sync.WaitGroup, jobs chan Job) {
	defer wg.Done()
	log.Println("contractor() started")

	wp := new(WorkersCount)

	go wp.run(ctx, jobs)

	for {
		select {
		case r, ok := <-wp.outbox():
			if !ok {
				continue
			}

			i, err := strconv.ParseInt(string(r.Descriptor.ID), 10, 64)
			if err != nil {
				log.Fatalf("unexpected error: %v", err)
			}

			val := r.Value.(int)
			log.Println("result: ", r)
			if val != int(i) {
				log.Fatalf("wrong value %v; expected %v", val, int(i)*2)
			}
		case <-ctx.Done():
			log.Println("context.Done() in Run()")
			return

		}
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result, id int) {
	defer wg.Done()
	log.Println("started worker ", id)

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				log.Println("iddle worker ", id)
				// return
			} else {
				results <- job.execute(ctx)
			}

		case <-ctx.Done():
			log.Println("context.Done() in worker ", id)
			results <- Result{
				Err: ctx.Err(),
			}
			return
		}
	}
}


func new(wcount int) WorkerPool {
	return WorkerPool{
		workersCount: wcount,
		results:      make(chan Result, wcount),
	}
}

func (wp WorkerPool) run(ctx context.Context, jobs chan Job) {
	var wg sync.WaitGroup
	log.Println("wp.Run() started")
	for i := 0; i < wp.workersCount; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobs, wp.results, i)
	}

	wg.Wait()
	close(wp.results)
}

func (wp WorkerPool) outbox() <-chan Result {
	return wp.results
}

type JobID string
type jobType string

type ExecutionFn func(ctx context.Context, args interface{}) (interface{}, error)

type JobDescriptor struct {
	ID    JobID
	JType jobType
}

type Result struct {
	Value      interface{}
	Err        error
	Descriptor JobDescriptor
}

type Job struct {
	Descriptor JobDescriptor
	ExecFn     ExecutionFn
	Args       interface{}
}

func (j Job) execute(ctx context.Context) Result {
	value, err := j.ExecFn(ctx, j.Args)
	if err != nil {
		return Result{
			Err:        err,
			Descriptor: j.Descriptor,
		}
	}

	return Result{
		Value:      value,
		Descriptor: j.Descriptor,
	}
}
