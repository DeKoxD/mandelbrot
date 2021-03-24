package queue

import (
	"errors"
)

type Queue struct {
	queue chan chan struct{}
	mx    chan struct{}
}

func (q *Queue) Begin() error {
	if len(q.queue) == cap(q.queue) {
		return errors.New("Queue full")
	}
	ch := make(chan struct{})
	q.queue <- ch

	// Lock
	<-q.mx

	headCh := <-q.queue
	close(headCh)

	<-ch
	return nil
}

func (q *Queue) End() {
	// Unlock
	q.mx <- struct{}{}
}

func NewQueue(size uint) Queue {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return Queue{
		queue: make(chan chan struct{}, size),
		mx:    ch,
	}
}
