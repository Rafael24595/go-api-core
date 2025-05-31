package routine

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

//TODO: Implement for request collection tests

type JobResult[T any] struct {
	Instance string
	Thread   int
	Output   T
}

type JobError[T any] struct {
	Instance string
	Thread   int
	Error    error
}

type Job[T any] func(ctx context.Context) (T, error)

type FactoryStreamPool[T any] struct {
	workerCount int
	bufferSize  int
	autoDrain   bool
	forceExit   bool
}

func AsyncStreamPool[T any](workerCount, bufferSize int) *FactoryStreamPool[T] {
	return &FactoryStreamPool[T]{
		workerCount: workerCount,
		bufferSize: bufferSize,
		autoDrain: false,
		forceExit: false,
	}
}

func SyncStreamPool[T any](bufferSize int) *FactoryStreamPool[T] {
	return &FactoryStreamPool[T]{
		workerCount: 1,
		bufferSize: bufferSize,
		autoDrain: false,
		forceExit: false,
	}
}

func (f *FactoryStreamPool[T]) EnableAutoDrain() *FactoryStreamPool[T] {
	f.autoDrain = true
	return f
}

func (f *FactoryStreamPool[T]) DisableAutoDrain() *FactoryStreamPool[T] {
	f.autoDrain = false
	return f
}

func (f *FactoryStreamPool[T]) GraceExit() *FactoryStreamPool[T] {
	f.forceExit = false
	return f
}

func (f *FactoryStreamPool[T]) ForceExit() *FactoryStreamPool[T] {
	f.forceExit = true
	return f
}

func (f *FactoryStreamPool[T]) Make() *StreamPool[T] {
	return newStreamPool[T](f.workerCount, f.bufferSize, f.autoDrain, f.forceExit)
}

type StreamPool[T any] struct {
	wgw         sync.WaitGroup
	wgj         sync.WaitGroup
	instance    string
	workerCount int
	autoDrain   bool
	forceExit   bool
	jobs        chan Job[T]
	results     chan JobResult[T]
	errors      chan JobError[T]
	ctx         context.Context
	cancel      context.CancelFunc
}

func newStreamPool[T any](workerCount, bufferSize int, autoDrain, forceExit bool) *StreamPool[T] {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan Job[T], bufferSize)
	results := make(chan JobResult[T], bufferSize)
	errors := make(chan JobError[T], bufferSize)

	pool := &StreamPool[T]{
		instance:    uuid.NewString(),
		workerCount: workerCount,
		autoDrain:   autoDrain,
		forceExit:   forceExit,
		jobs:        jobs,
		results:     results,
		errors:      errors,
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool.start()
}

func (p *StreamPool[T]) worker(thread int) *StreamPool[T] {
	defer p.wgw.Done()
	for {
		select {
		case <-p.ctx.Done():
			return p.drainJobs()
		case job, ok := <-p.jobs:
			if !ok {
				return p
			}

			output, err := job(p.ctx)
			p.wgj.Done()
			if err != nil {
				p.manageError(thread, err)
				continue
			}
			p.manageOutput(thread, output)
		}
	}
}

func (p *StreamPool[T]) manageError(thread int, err error) *StreamPool[T] {
	error := JobError[T]{
		Instance: p.instance,
		Thread:   thread,
		Error:    err,
	}

	select {
	case p.errors <- error:
	default:
		if !p.autoDrain {
			message := fmt.Sprintf(
				"[StreamPool:%s] Errors overflow on thread %d: unable to submit error — consider increasing buffer size or enabling auto-drain mode.",
				p.instance,
				thread,
			)
			panic(message)
		}
		p.drainErrors()
	}

	return p
}

func (p *StreamPool[T]) manageOutput(thread int, output T) *StreamPool[T] {
	result := JobResult[T]{
		Instance: p.instance,
		Thread:   thread,
		Output:   output,
	}

	select {
	case p.results <- result:
	default:
		if !p.autoDrain {
			message := fmt.Sprintf(
				"[StreamPool:%s] Results overflow on thread %d: unable to submit result — consider increasing buffer size or enabling auto-drain mode.",
				p.instance,
				thread,
			)
			panic(message)
		}
		p.drainResults()
	}

	return p
}

func (p *StreamPool[T]) start() *StreamPool[T] {
	p.wgw.Add(p.workerCount)
	for i := range p.workerCount {
		go p.worker(i)
	}
	return p
}

func (p *StreamPool[T]) Id(job Job[T]) string {
	return p.instance
}

func (p *StreamPool[T]) Submit(job Job[T]) bool {
	p.wgj.Add(1)
	select {
	case <-p.ctx.Done():
		p.wgj.Done()
		return false
	case p.jobs <- job:
		return true
	default:
		p.wgj.Done()
		return false
	}
}

func (p *StreamPool[T]) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		p.wgj.Wait()
		close(done)
	}()
	return done
}

func (p *StreamPool[T]) Results() <-chan JobResult[T] {
	return p.results
}

func (p *StreamPool[T]) Errors() <-chan JobError[T] {
	return p.errors
}

func (p *StreamPool[T]) Stop() *StreamPool[T] {
	p.cancel()
	close(p.jobs)
	p.wgw.Wait()
	if !p.forceExit && len(p.jobs) > 0 {
		p.wgj.Wait()
	}
	close(p.results)
	close(p.errors)
	return p
}

func (p *StreamPool[T]) drainJobs() *StreamPool[T] {
	for {
		select {
		case _, ok := <-p.jobs:
			if !ok {
				return p
			}
			p.wgj.Done()
		default:
			return p
		}
	}
}

func (p *StreamPool[T]) drainErrors() *StreamPool[T] {
	go func() {
		for range p.Errors() {
		}
	}()
	return p
}

func (p *StreamPool[T]) drainResults() *StreamPool[T] {
	go func() {
		for range p.Results() {
		}
	}()
	return p
}
