package goxeca

import (
	"container/list"
	"sync"
)

type Queue struct {
	queue *list.List
	mu    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		queue: list.New(),
	}
}

func (q *Queue) Enqueue(job *Job) {
	q.mu.Lock()

	defer q.mu.Unlock()
	q.queue.PushBack(job)
}

func (q *Queue) Dequeue() *Job {
	q.mu.Lock()

	defer q.mu.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	element := q.queue.Front()
	q.queue.Remove(element)
	return element.Value.(*Job)
}

func (q *Queue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.queue.Len()
}
