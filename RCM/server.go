package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	// "net"
	"sync"
)

type WitnessEntry struct {
	id      int
	counter int
}

type Server struct {
	update_needed  chan bool
	m_data         map[int]string
	m_data_lock    sync.RWMutex
	vec_clock      []int
	vec_clock_lock sync.Mutex
	vec_clock_cond *sync.Cond
	queue          Queue
	witness        map[WitnessEntry]int
	witness_lock   sync.Mutex
	publisher_lock sync.Mutex
	publisher      *zmq.Socket
	subscriber     *zmq.Socket
}

func (svr *Server) init(pub_port string) {
	svr.update_needed = make(chan bool, 100)
	// init data as key(int)-value(string) pair
	svr.m_data = make(map[int]string)
	svr.m_data_lock = sync.RWMutex{}
	// init vector timestamp with length group_size
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
	svr.witness = make(map[WitnessEntry]int)
	svr.witness_lock = sync.Mutex{}
	svr.publisher = createPublisherSocket(pub_port)
	svr.subscriber = createSubscriberSocket()
	go svr.subscribe()
}

// Actions to take if server receives READ message
func (svr *Server) recvRead(key int, id int, counter int, vec_i []int) *Message {
	// wait until t_server is greater than t_i
	svr.vec_clock_cond.L.Lock()
	for !smallerEqualExceptI(vec_i, svr.vec_clock, 999999) {
		svr.vec_clock_cond.Wait()
	}
	svr.vec_clock_cond.L.Unlock()
	// send RESP message to client i
	svr.m_data_lock.RLock()
	msg := Message{Kind: RESP, Counter: counter, Val: svr.m_data[key], Vec: svr.vec_clock}
	svr.m_data_lock.RUnlock()
	return &msg
	//send(&msg, mem_list[id])
}

// Actions to take if server receives WRITE message
func (svr *Server) recvWrite(key int, val string, id int, counter int, vec_i []int) *Message{
	// broadcast UPDATE message
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
	svr.publish(&msg)
	fmt.Printf("Server %d published msg UPDATE\n", node_id)

	// wait until t_server is greater than t_i
	svr.vec_clock_cond.L.Lock()
	for !smallerEqualExceptI(vec_i, svr.vec_clock, 999999) {
		fmt.Println(vec_i)
		fmt.Println(svr.vec_clock)
		svr.vec_clock_cond.Wait()
	}
	svr.vec_clock_cond.L.Unlock()
	// send ACK message to client i
	msg = Message{Kind: ACK, Counter: counter, Vec: svr.vec_clock}
	return &msg
	//send(&msg, mem_list[id])
}

// Actions to take if server receives UPDATE message
func (svr *Server) recvUpdate(key int, val string, id int, counter int, vec_i []int) {
	entry := WitnessEntry{id: id, counter: counter}
	svr.witness_lock.Lock()
	if _, isIn := svr.witness[entry]; isIn {
		svr.witness[entry] += 1
	} else {
		svr.witness[entry] = 1
	}
	witness_num := svr.witness[entry]
	svr.witness_lock.Unlock()

	if witness_num == 1 {
		msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
		svr.publish(&msg)
		fmt.Printf("Server %d published msg UPDATE\n", node_id)
	}
	if witness_num == F+1 {
		queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
		svr.queue.Enqueue(queue_entry)
		svr.update_needed <- true
		// fmt.Println("server enqueues entry: ", queue_entry)
	}
}

// // Server listener
// func (svr *Server) recv() {
// 	// resolve for udp address by membership list and id
// 	udpAddr, err1 := net.ResolveUDPAddr("udp4", mem_list[id])
// 	if err1 != nil {
// 		fmt.Println("address not found")
// 	}

// 	// create listner socket by address
// 	conn, err2 := net.ListenUDP("udp", udpAddr)
// 	if err2 != nil {
// 		fmt.Println("address can't listen")
// 	}
// 	defer conn.Close()

// 	for status {
// 		c := make(chan Message)

// 		go func() {
// 			//buffer size is 1024 bytes
// 			buf := make([]byte, 1024)
// 			num, _, err3 := conn.ReadFromUDP(buf)
// 			if err3 != nil {
// 				fmt.Println(err3)
// 			}
// 			//deserialize the received data and output to channel
// 			c <- getMsgFromGob(buf[:num])
// 		}()

// 		msg := <-c

// 		switch msg.Kind {
// 		case READ:
// 			// fmt.Println("server receives READ message with vec_clock", msg.Vec)
// 			go svr.recvRead(msg.Key, msg.Id, msg.Counter, msg.Vec)
// 		case WRITE:
// 			// fmt.Println("server receives WRITE message with vec_clock", msg.Vec)
// 			go svr.recvWrite(msg.Key, msg.Val, msg.Id, msg.Counter, msg.Vec)
// 		case UPDATE:
// 			// fmt.Println("server receives UPDATE message with vec_clock", msg.Vec)
// 			go svr.recvUpdate(msg.Key, msg.Val, msg.Id, msg.Counter, msg.Vec)
// 		}
// 	}
// }

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
