package main

import (
	"fmt"
	"sync"
	zmq "github.com/pebbe/zmq4"
)

type WitnessEntry struct {
	key 	int
	val 	string
	id      int
	counter int
}

type Server struct {
	m_data			map[int]string
	m_data_lock		sync.RWMutex
	vec_clock		[]int
	vec_clock_lock	sync.RWMutex
	vec_clock_cond	*sync.Cond
	queue			Queue
	witness 		map[WitnessEntry]map[int]bool
	has_sent		map[WitnessEntry]bool
	has_sent_lock 	sync.Mutex
	publisher_lock sync.Mutex
	publisher      *zmq.Socket
	subscriber     *zmq.Socket
}

func (svr *Server) init(pub_port string) {
	// init data as key(int)-value(string) pair
	svr.m_data = make(map[int] string)
	svr.m_data_lock = sync.RWMutex{}
	// init vector timestamp with length group_size
	svr.vec_clock = make([]int, NUM_CLIENT)
	svr.vec_clock_lock = sync.RWMutex{}
	svr.vec_clock_cond = sync.NewCond(&svr.vec_clock_lock)
	// set vector timestamp to zero
	for i:= 0; i < NUM_CLIENT; i++ {
		svr.vec_clock[i] = 0
	}
	// init queue
	svr.queue.Init()
	// init witness
	svr.witness = make(map[WitnessEntry]map[int]bool)
	svr.has_sent = make(map[WitnessEntry]bool)
	svr.has_sent_lock = sync.Mutex{}
	svr.publisher = createPublisherSocket(pub_port)
	svr.subscriber = createSubscriberSocket()
	go svr.subscribe()
}

func (svr *Server) recvRead(key int, id int, counter int, vec_i []int) *Message {
	svr.m_data_lock.RLock()
	msg := Message{Kind: RESP, Counter: counter, Key: key, Val: svr.m_data[key], Vec: svr.vec_clock}
	svr.m_data_lock.RUnlock()
	return &msg
}

func (svr *Server) recvWrite(key int, val string, id int, counter int, vec_i []int) *Message {
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
	entry := WitnessEntry{id: id, counter: counter}
	svr.has_sent_lock.Lock()
	if _,isIn := svr.has_sent[entry]; !isIn {
		svr.publish(&msg)
		svr.has_sent[entry] = true
		fmt.Printf("Server %d published msg UPDATE in response to WRITE from client %d\n", node_id, id)
	}
	svr.has_sent_lock.Unlock()

	msg = Message{Kind: ACK, Counter: counter, Vec: svr.vec_clock}
	return &msg
}

// Actions to take if server receives UPDATE message
func (svr *Server) recvUpdate(key int, val string, id int, counter int, vec_i []int, sender_id int) {
	entry := WitnessEntry{id: id, counter: counter}
	if _,isIn := svr.witness[entry]; isIn {
		if _,hasReceived := svr.witness[entry][sender_id]; !hasReceived {
			svr.witness[entry][sender_id] = true

			// Publish UPDATE
			svr.has_sent_lock.Lock()
			if _,isIn := svr.has_sent[entry]; !isIn {
				msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
				svr.publish(&msg)
				svr.has_sent[entry] = true
				fmt.Printf("Server %d published msg UPDATE in response to UPDATE from server %d\n", node_id, sender_id)
			}
			svr.has_sent_lock.Unlock()

			if len(svr.witness[entry]) == F+1 {
				queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
				svr.queue.Enqueue(queue_entry)
				go svr.update()
				// fmt.Println("server enqueues entry: ", queue_entry)
			}
		}
	} else {
		svr.witness[entry] = make(map[int]bool)
		svr.witness[entry][sender_id] = true

		// Publish UPDATE
		svr.has_sent_lock.Lock()
		if _,isIn := svr.has_sent[entry]; !isIn {
			msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
			svr.publish(&msg)
			svr.has_sent[entry] = true
			fmt.Printf("Server %d published msg UPDATE in response to UPDATE from server %d\n", node_id, sender_id)
		}
		svr.has_sent_lock.Unlock()

		if len(svr.witness[entry]) == F+1 {
			queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
			svr.queue.Enqueue(queue_entry)
			go svr.update()
			// fmt.Println("server enqueues entry: ", queue_entry)
		}
	}
}

// infinitely often update the local storage
func (svr *Server) update(){
	msg := svr.queue.Dequeue()
	if msg != nil {
		// fmt.Println("server receives msg with vec_clock: ", msg.Vec)
		// fmt.Println("server has vec_clock: ", svr.vec_clock)
		svr.vec_clock_cond.L.Lock()
		for svr.vec_clock[msg.Id] != msg.Vec[msg.Id]-1 || !smallerEqualExceptI(msg.Vec, svr.vec_clock, msg.Id) {
			svr.vec_clock_cond.Wait()
			if svr.vec_clock[msg.Id] > msg.Vec[msg.Id]-1 {
				return
			}
		}
		// update timestamp and write to local memory
		svr.vec_clock[msg.Id] = msg.Vec[msg.Id]
		// fmt.Println("server increments vec_clock: ", svr.vec_clock)
		svr.vec_clock_cond.Broadcast()
		svr.vec_clock_cond.L.Unlock()
		svr.m_data_lock.Lock()
		svr.m_data[msg.Key] = msg.Val
		svr.m_data_lock.Unlock()
	}
}

// helper function that return true if elements of vec1 are smaller than those of vec2 except i-th element; false otherwise
func smallerEqualExceptI(vec1 []int, vec2 []int, i int) bool {
	if len(vec1) != len(vec2) {
		panic("vector clocks are of different lengths")
	}
	for index:=0; index<len(vec1); index++ {
		if index == i {
			continue
		}
		if vec1[index] > vec2[index] {
			return false
		}
	}
	return true
}