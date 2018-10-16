package main

import (
	// "container/heap"
	// "time"
)

type Node struct {
	m_data    map[int]string
	vec_clock []int
	// outQueue  PriorityQueue
	// inQueue   PriorityQueue
	outQueue  chan Message
	inQueue   chan Message
}

// initialize the node
func (n *Node) init(group_size int) {
	// init data as key(int)-value(string) pair
	n.m_data = make(map[int]string)
	// init vector timestamp with length group_size
	n.vec_clock = make([]int, group_size)
	// set vector timestamp to zero
	for i := 0; i < group_size; i++ {
		n.vec_clock[i] = 0
	}
	// // init priority queue outQueue
	// n.outQueue = make(PriorityQueue, 0)
	// heap.Init(&n.outQueue)
	// // init priority queue inQueue
	// n.inQueue = make(PriorityQueue, 0)
	// heap.Init(&n.inQueue)
	n.outQueue = make(chan Message, 10000)
	n.inQueue = make(chan Message, 10000)
}

// perform read(key int), return value string
func (n *Node) read(key int) string {
	mutex.Lock()
	res := n.m_data[key]
	mutex.Unlock()
	return res
}

// perform write(id int, key int, value string)
func (n *Node) write(key int, value string) {
	// update vector clock
	n.vec_clock[id] += 1
	// update key-value pair
	mutex.Lock()
	n.m_data[key] = value
	mutex.Unlock()
	// create Message object and push to outQueue
	msg := Message{Type: SERVER, Id: id, Key: key, Val: value, Vec: n.vec_clock}
	// heap.Push(&n.outQueue, &msg)
	n.outQueue <- msg
}

// apply action
func (n *Node) apply() {
	for status {
		// while inqueue is not empty, compare it and update
		// for n.inQueue.Len() > 0 {
		// pop
		// msg := *heap.Pop(&n.inQueue).(*Message)
		select {
		case msg := <- n.inQueue:
			if n.compareTo(msg.Id, msg.Vec) {
				// update local vector clock
				n.vec_clock[msg.Id] = msg.Vec[msg.Id]
				// update memory
				mutex.Lock()
				n.m_data[msg.Key] = msg.Val
				mutex.Unlock()
			} else {
				// heap.Push(&n.inQueue, &msg)
				n.inQueue <- msg
			}
		}
		// }
		// wait for inqueue to be non-empty
		// time.Sleep(10 * time.Millisecond)
	}
}

// send action
func (n *Node) send() {
	for status {
		// while out queue is not empty, pop out msg and broadcast it
		// for n.outQueue.Len() > 0 {
		// pop
		// msg := heap.Pop(&n.outQueue).(*Message)
		select {
		case msg := <- n.outQueue:
			// broadcast
			broadcast(&msg)
		}
		// }
		// wait for outqueue to be non-empty
		// time.Sleep(10 * time.Millisecond)
	}
}

// helper function for apply action
func (n *Node) compareTo(id int, vec []int) bool {
	flag := true
	for i := 0; i < len(vec); i++ {
		// new_vec[k] leq local_vec[k] for k neq id
		if i != id && vec[i] > n.vec_clock[i] {
			flag = false
			break
			// new_vec[id] = local_vec[id] + 1
		} else if i == id && vec[i] != n.vec_clock[i]+1 {
			flag = false
			break
		}
	}
	return flag
}
