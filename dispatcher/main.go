package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// create the dispatcher
	dispatcher := &dispatcher{}

	// create global quit channel
	globalQuit := make(chan struct{})

	// create the workers
	for i := 1; i < 6; i++ {
		w := &worker{
			id:   i,
			quit: globalQuit,
		}
		dispatcher.Push(w)
		w.Start()
	}

	time.Sleep(time.Second)

	// start message generator
	go func() {
		counter := 0
		for {
			dispatcher.Iterate(
				func(w *worker) {
					w.source <- fmt.Sprintf("message no. %d", counter)
				},
			)
			counter++
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Use a dirty hack to avoid everything ending too early
	fmt.Println("waiting 5 sec before ending main go-routine")
	time.Sleep(5 * time.Second)
}

// dispatcher

type dispatcher struct {
	sync.Mutex
	workers []*worker
}

func (s *dispatcher) Push(w *worker) {
	fmt.Println(fmt.Sprintf("Pushing worker %d to list", w.id))
	s.Lock()
	defer s.Unlock()
	s.workers = append(s.workers, w)
}

func (s *dispatcher) Iterate(routine func(*worker)) {
	s.Lock()
	defer s.Unlock()
	for _, worker := range s.workers {
		routine(worker)
	}
}

// worker

type worker struct {
	id     int
	source chan interface{}
	quit   chan struct{}
}

func (w *worker) Start() {
	fmt.Println(fmt.Sprintf("Start worker %d", w.id))
	w.source = make(chan interface{}, 10)
	go func() {
		for {
			select {
			case msg := <-w.source:
				fmt.Println(fmt.Sprintf("worker %d received %s", w.id, msg))
			case <-w.quit:
				return
			}
		}
	}()
}
