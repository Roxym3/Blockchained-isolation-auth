package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Task struct {
	Handler func() error
}

func NewTask(h func() error) *Task {
	return &Task{Handler: h}
}

type TaskPool struct {
	capacity int32
	running  int32
	JobCh    chan *Task
	mut      sync.Mutex
	wg       sync.WaitGroup
	closed   bool
}

func NewPool(cap int32) *TaskPool {
	return &TaskPool{
		capacity: cap,
		JobCh:    make(chan *Task, cap*2),
	}
}

func (p *TaskPool) AddTask(t *Task) error {
	p.mut.Lock()
	if p.closed {
		p.mut.Unlock()
		return fmt.Errorf("pool is closed")
	}
	p.wg.Add(1)
	if atomic.LoadInt32(&p.running) < p.capacity {
		p.run()
	}
	p.mut.Unlock()

	p.JobCh <- t
	return nil
}

func (p *TaskPool) run() {
	atomic.AddInt32(&p.running, 1)
	go func() {
		defer atomic.AddInt32(&p.running, -1)

		for task := range p.JobCh {
			func() {
				defer p.wg.Done()
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("panic recovered:%v\n", err)
					}
				}()

				if err := task.Handler(); err != nil {
					fmt.Printf("task failed:%v\n", err)
				}
			}()
		}
	}()
}

func (p *TaskPool) Wait() {
	p.mut.Lock()
	p.closed = true
	p.mut.Unlock()

	close(p.JobCh)

	p.wg.Wait()
}
