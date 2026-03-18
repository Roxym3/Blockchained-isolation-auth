package pkg

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
	Mut      *sync.Mutex
}

func NewPool(cap int32) *TaskPool {
	return &TaskPool{
		capacity: cap,
		JobCh:    make(chan *Task, cap*2),
	}
}

func (p *TaskPool) AddTask(t *Task) {
	p.Mut.Lock()
	if atomic.LoadInt32(&p.running) < p.capacity {
		p.run()
	}
	p.Mut.Unlock()

	p.JobCh <- t
}

func (p *TaskPool) run() {
	atomic.AddInt32(&p.running, 1)
	go func() {
		defer func() {
			atomic.AddInt32(&p.running, -1)
			if err := recover(); err != nil {
				fmt.Printf("panic recovered:%v\n", err)
			}
		}()

		for task := range p.JobCh {
			if err := task.Handler(); err != nil {
				fmt.Printf("task failed:%v", err)
			}
		}
	}()
}
