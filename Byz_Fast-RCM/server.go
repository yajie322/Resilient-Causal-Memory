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
	update_needed 	chan bool
	m_data			map[int]string
	m_data_lock		sync.RWMutex
	vec_clock		[]int
	vec_clock_lock	sync.RWMutex
	vec_clock_cond	*sync.Cond
	queue			Queue
	witness 		map[WitnessEntry]int
	publisher_lock sync.Mutex
	publisher      *zmq.Socket
	subscriber     *zmq.Socket
}

func (svr *Server) init(pub_port string) {
	svr.update_needed = make(chan bool, 100)
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
	svr.witness = make(map[WitnessEntry] int)
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
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
	svr.publish(&msg)
	// fmt.Printf("Server %d published msg UPDATE\n", node_id)

	msg = Message{Kind: ACK, Counter: counter, Vec: svr.vec_clock}
	return &msg
}

// Actions to take if server receives UPDATE message
func (svr *Server) recvUpdate(key int, val string, id int, counter int, vec_i []int) {
	entry := WitnessEntry{key: key, val: val, id: id, counter: counter}
	if _, isIn := svr.witness[entry]; isIn {
		svr.witness[entry] += 1
	} else {
		svr.witness[entry] = 1
	}
	witness_num := svr.witness[entry]

	if witness_num == F+1 {
		msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
		svr.publish(&msg)
		queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
		svr.queue.Enqueue(queue_entry)
		svr.update_needed <- true
		// fmt.Println("server enqueues entry: ", queue_entry)
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