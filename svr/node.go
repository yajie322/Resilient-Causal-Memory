package main

import (
	// "time"
	"container/heap"
	"time"
	zmq "github.com/pebbe/zmq4"
)

type Server struct {
	mData    map[int]string
	vecClock []int
	inQueue   PriorityQueue
	outQueue  chan Message
	publisher *zmq.Socket
	subscriber *zmq.Socket
}

// initialize the node
func (svr *Server) init(group_size int) {
	// init data as key(int)-value(string) pair
	svr.mData = make(map[int]string)
	// init vector timestamp with length group_size
	svr.vecClock = make([]int, group_size)
	// set vector timestamp to zero
	for i := 0; i < group_size; i++ {
		svr.vecClock[i] = 0
	}
	// init queue outQueue
	svr.outQueue = make(chan Message, 10000)
	// init priority queue inQueue
	svr.inQueue = make(PriorityQueue, 0)
	heap.Init(&svr.inQueue)
	// init sockets
	svr.publisher = createPublisherSocket(svr_list[node_id])
	svr.subscriber = createSubscriberSocket()
	go svr.recv()
}

// perform read(key int), return value string
func (svr *Server) read(key int) string {
	mutex.Lock()
	res := svr.mData[key]
	mutex.Unlock()
	return res
}

// perform write(id int, key int, value string)
func (svr *Server) write(key int, value string) {
	// update vector clock
	svr.vecClock[node_id] += 1
	// update key-value pair
	mutex.Lock()
	svr.mData[key] = value
	mutex.Unlock()
	// create Message object and push to outQueue
	msg := Message{Type: ACK, Id: node_id, Key: key, Val: value, Vec: svr.vecClock}
	// heap.Push(&svr.outQueue, &msg)
	svr.outQueue <- msg
}

func (svr *Server) recv() {
	for status{
		//	get bytes
		b,_ := svr.subscriber.RecvMessageBytes(0)
		msg := getMsgFromGob(b[0])
		heap.Push(&svr.inQueue, &msg)
	}
}

// send action
func (svr *Server) send() {
	for status {
		// while out queue is not empty, pop out msg and broadcast it
		select {
		case msg := <- svr.outQueue:
			// broadcast
			svr.publish(&msg)
		}
	}
}

// apply action
func (svr *Server) apply() {
	for status {
		// while inqueue is not empty, compare it and update
		for svr.inQueue.Len() > 0 {
			// peek
			msg := svr.inQueue.Peek()
			if svr.compareTo(msg.Id, msg.Vec) {
				// pop
				msg = heap.Pop(&svr.inQueue).(*Message)
				// update local vector clock
				svr.vecClock[msg.Id] = msg.Vec[msg.Id]
				// update memory
				mutex.Lock()
				svr.mData[msg.Key] = msg.Val
				mutex.Unlock()
			}
		}
		// wait for inqueue to be non-empty
		time.Sleep(time.Millisecond)
	}
}

// helper function for apply action
func (svr *Server) compareTo(id int, vec []int) bool {
	for i := 0; i < len(vec); i++ {
		if i != id && vec[i] > svr.vecClock[i] {
			// new_vec[k] leq local_vec[k] for k neq id
			return false
		} else if i == id && vec[i] != svr.vecClock[i]+1 {
			// new_vec[id] = local_vec[id] + 1
			return false
		}
	}
	return true
}
