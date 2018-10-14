package main

import (
	"time"
	"net"
	"fmt"
)

type WitnessEntry struct {
	id 		int
	counter int
}

type Server struct {
	m_data		map[int]string
	vec_clock	[]int
	queue		Queue
	witness 	map[WitnessEntry]int
}

func (svr *Server) init(group_size int) {
	// init data as key(int)-value(string) pair
	svr.m_data = make(map[int] string)
	// init vector timestamp with length group_size
	svr.vec_clock = make([]int, group_size)
	// set vector timestamp to zero
	for i:= 0; i < group_size; i++ {
		svr.vec_clock[i] = 0
	}
	// init queue
	svr.queue.Init()
	// init witness
	svr.witness = make(map[WitnessEntry] int)
}

func (svr *Server) recvRead(key int, id int, counter int, vec_i []int){
	msg := Message{Kind: RESP, Counter: counter, Val: svr.m_data[key], Vec: svr.vec_clock}
	send(&msg, mem_list[id])
}

func (svr *Server) recvWrite(key int, val string, id int, counter int, vec_i []int) {
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
	broadcast(&msg)
	msg = Message{Kind: ACK, Counter: counter, Vec: svr.vec_clock}
	send(&msg, mem_list[id])
}

// Actions to take if server receives UPDATE message
func (svr *Server) recvUpdate(key int, val string, id int, counter int, vec_i []int) {
	entry := WitnessEntry{id: id, counter: counter}
	if _, isIn := svr.witness[entry]; isIn {
		svr.witness[entry] += 1
	} else {
		svr.witness[entry] = 1
	}
	if svr.witness[entry] == 1 {
		msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
		broadcast(&msg)
	}
	if svr.witness[entry] == F+1 {
		queue_entry := QueueEntry{Key: key, Val: val, Id: id, Vec: vec_i}
		svr.queue.Enqueue(queue_entry)
	}
}

func (svr *Server) recv(){
	// resolve for udp address by membership list and id
	udpAddr,err1 := net.ResolveUDPAddr("udp4", mem_list[id])
	if err1 != nil {
		fmt.Println("address not found")
	}

	// create listner socket by address
	conn,err2 := net.ListenUDP("udp", udpAddr)
	if err2 != nil {
		fmt.Println("address can't listen")
	}
	defer conn.Close()

	for status {
		c := make(chan Message)

		go func() {
			//buffer size is 1024 bytes
			buf := make([]byte, 1024)
			num,_,err3 := conn.ReadFromUDP(buf)
			if err3 != nil {
				fmt.Println(err3)
			}
			//deserialize the received data and output to channel
			c <- getMsgFromGob(buf[:num])
		}()

		msg := <-c
		
		switch msg.Kind {
			case READ:
				go svr.recvRead(msg.Key, msg.Id, msg.Counter, msg.Vec)
			case WRITE:
				go svr.recvWrite(msg.Key, msg.Val, msg.Id, msg.Counter, msg.Vec)
			case UPDATE:
				go svr.recvUpdate(msg.Key, msg.Val, msg.Id, msg.Counter, msg.Vec)
		}
	}
}

// infinitely often update the local storage
func (svr *Server) update(){
	for svr.queue.Len() > 0 {
		msg := svr.queue.Dequeue()
		// wait until it's time to update
		fmt.Println(msg.Vec)
		fmt.Println(svr.vec_clock)
		for svr.vec_clock[msg.Id] < msg.Vec[msg.Id]-1 || !smallerEqualExceptI(msg.Vec, svr.vec_clock, msg.Id) {
			time.Sleep(time.Millisecond)
		}
		// update timestamp and write to local memory
		svr.vec_clock[msg.Id] = msg.Vec[msg.Id]
		svr.m_data[msg.Key] = msg.Val
		fmt.Println("server ", id, " has m_data: ", svr.m_data)
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