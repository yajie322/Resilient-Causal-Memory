package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"sync"
)

type WitnessEntry struct {
	key 	string
	val 	string
	id      int
	counter int
}

type Server struct {
	vec_clock      []int
	vec_clock_lock sync.Mutex
	vec_clock_cond *sync.Cond
	queue          Queue
	witness        map[WitnessEntry]map[int]bool
	has_sent		map[WitnessEntry]bool
	has_sent_lock	sync.Mutex
	publisher_lock sync.Mutex
	publisher      *zmq.Socket
	subscriber     *zmq.Socket
}

func (svr *Server) init(pub_port string) {
	svr.vec_clock = make([]int, NUM_CLIENT)
	svr.vec_clock_lock = sync.Mutex{}
	svr.vec_clock_cond = sync.NewCond(&svr.vec_clock_lock)
	// set vector timestamp to zero
	for i := 0; i < NUM_CLIENT; i++ {
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

// Actions to take if server receives READ message
func (svr *Server) recvRead(key string, id int, counter int, vec_i []int) *Message {
	// wait until t_server is greater than t_i
	svr.waitUntilServerClockGreaterExceptI(vec_i, 999999)

	// send RESP message to client i
	msg := Message{Kind: RESP, Counter: counter, Val: d.ReadString(key), Vec: svr.vec_clock}
	return &msg
}

// Actions to take if server receives WRITE message
func (svr *Server) recvWrite(key string, val string, id int, counter int, vec_i []int) *Message{
	// broadcast UPDATE message
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
	entry := WitnessEntry{id: id, counter: counter}
	svr.has_sent_lock.Lock()
	if _,isIn := svr.has_sent[entry]; !isIn {
		svr.publish(&msg)
		svr.has_sent[entry] = true
		fmt.Printf("Server %d published msg UPDATE in response to WRITE from client %d\n", node_id, id)
	}
	svr.has_sent_lock.Unlock()

	// wait until t_server is greater than t_i
	svr.waitUntilServerClockGreaterExceptI(vec_i, 999999)

	// send ACK message to client i
	msg = Message{Kind: ACK, Counter: counter, Vec: svr.vec_clock}
	return &msg
}

// Actions to take if server receives UPDATE message
func (svr *Server) recvUpdate(key string, val string, id int, counter int, vec_i []int, sender_id int) {
	entry := WitnessEntry{key: key, val: val, id: id, counter: counter}

	if _, isIn := svr.witness[entry]; isIn {
		if _, hasReceived := svr.witness[entry][sender_id]; !hasReceived {
			svr.witness[entry][sender_id] = true

			if len(svr.witness[entry]) == F+1 {
				// Publish UPDATE
				svr.has_sent_lock.Lock()
				if _,isIn := svr.has_sent[entry]; !isIn {
					msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
					svr.publish(&msg)
					svr.has_sent[entry] = true
					fmt.Printf("Server %d published msg UPDATE in response to UPDATE from server %d\n", node_id, sender_id)
				}
				svr.has_sent_lock.Unlock()

				queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
				svr.queue.Enqueue(queue_entry)
				go svr.update()
				// fmt.Println("server enqueues entry: ", queue_entry)
			}
		}
	} else {
		svr.witness[entry] = make(map[int]bool)
		svr.witness[entry][sender_id] = true

		if len(svr.witness[entry]) == F+1 {
			// Publish UPDATE
			svr.has_sent_lock.Lock()
			if _,isIn := svr.has_sent[entry]; !isIn {
				msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i, Sender: node_id}
				svr.publish(&msg)
				svr.has_sent[entry] = true
				fmt.Printf("Server %d published msg UPDATE in response to UPDATE from server %d\n", node_id, sender_id)
			}
			svr.has_sent_lock.Unlock()

			queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
			svr.queue.Enqueue(queue_entry)
			go svr.update()
			// fmt.Println("server enqueues entry: ", queue_entry)
		}
	}
}

// infinitely often update the local storage
func (svr *Server) update() {
	msg := svr.queue.Dequeue()
	if msg != nil {
		// fmt.Println("server receives msg with vec_clock: ", msg.Vec)
		// fmt.Println("server has queue: ", svr.queue.values)
		// fmt.Println("server has vec_clock: ", svr.vec_clock)
		svr.vec_clock_cond.L.Lock()
		for svr.vec_clock[msg.Id] != msg.Vec[msg.Id]-1 || !smallerEqualExceptI(msg.Vec, svr.vec_clock, msg.Id) {
			if svr.vec_clock[msg.Id] > msg.Vec[msg.Id]-1 {
				return
			}
			svr.vec_clock_cond.Wait()
		}
		// update timestamp and write to local memory
		svr.vec_clock[msg.Id] = msg.Vec[msg.Id]
		// fmt.Println("server increments vec_clock: ", svr.vec_clock)
		svr.vec_clock_cond.Broadcast()
		svr.vec_clock_cond.L.Unlock()

		if err := d.WriteString(msg.Key,msg.Val); err != nil{
			fmt.Println(err)
		}
	}
}

// helper function that return true if elements of vec1 are smaller than those of vec2 except i-th element; false otherwise
func smallerEqualExceptI(vec1 []int, vec2 []int, i int) bool {
	if len(vec1) != len(vec2) {
		panic("vector clocks are of different lengths")
	}
	for index := 0; index < len(vec1); index++ {
		if index == i {
			continue
		}
		if vec1[index] > vec2[index] {
			return false
		}
	}
	return true
}

func (svr *Server) waitUntilServerClockGreaterExceptI(vec []int, i int) {
	svr.vec_clock_cond.L.Lock()
	for !smallerEqualExceptI(vec, svr.vec_clock, i) {
		svr.vec_clock_cond.Wait()
	}
	svr.vec_clock_cond.L.Unlock()
}
