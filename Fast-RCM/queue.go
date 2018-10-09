package main

import {
	"sync"
}

type QueueEntry struct {
	Key		int
	Val 	int
	Id 		int
	Vec		[]int
}

type Queue struct {
	lock	*sync.Mutex
	values	[]QueueEntry
}

func (q *Queue) Init() {
	q.lock = &sync.Mutex{}
	q.values = make([]QueueEntry, 0)
}

func (q *Queue) Enqueue(entry QueueEntry) {
  for {
    q.lock.Lock()
    q.values = append(q.values, entry)
    q.lock.Unlock()
    return
  }
}

func (q *Queue) Dequeue() *QueueEntry {
  for {
    if (len(q.values) > 0) {
      q.lock.Lock()
      entry := q.values[0]
      q.values = q.values[1:]
      q.lock.Unlock()
      return &entry
    }
    return nil
  }
  return nil
}

func (q *Queue) Len() int {
	return len(q.values)
}