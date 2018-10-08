package main

import {
	"time"
	//"github.com/golang-collections/go-datastructures/queue"	
}

type Server struct {
	m_data		map[int]string
	vec_clock	[]int
	queue		PriorityQueue
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

}

func (svr *Server) recvRead(key int, id int, counter int, vec_i []int){
	for !smallerEqualExceptI(vec_i, svr.vec_clock, 999999) {
		time.Sleep(time.Millisecond)
	}
	msg := Message{Kind: RESP, Counter: counter, M_data: svr.m_data, Vec: svr.vec_clock}
	send(msg, mem_list[id])
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

		select {
		case msg := <-c:
			// push the message to inQueue
			heap.Push(&n.inQueue, &msg)
		}
}


func (svr *Server) update(){
	for svr.queue.Len() > 0 {
		msg := svr.queue[0] // This might need to be changed depending on the implementation of our queue
		for svr.vec_clock[msg.Id] != msg.Vec[msg.Id]-1 || smallerEqualExceptI(msg.Vec, svr.vec_clock, msg.Id) {
			time.Sleep(time.Millisecond)
		}
		// update timestamp and write to local memory
		svr.vec_clock[msg.Id] = msg.Vec[msg.Id]
		svr.m_data[msg.Key] = msg.Val
	}
}

// helper function that return true if elements of vec1 are smaller than those of vec2 except i-th element; false otherwise
func smallerEqualExceptI(vec1 []int, vec2 []int, i int) bool {
	if len(vec1) != len(vec2) {
		panic()
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