package main

import {
	"time"
}

type Server struct {
	m_data		map[int]string
	vec_clock	[]int
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
	msg := Message{Kind: RESP, Counter: counter, Val: svr.m_data[key], Vec: svr.vec_clock}
	send(msg, mem_list[id])
}

func (svr *Server) recvWrite(key int, val string, id int, counter int, vec_i []int) {
	msg := Message{Kind: UPDATE, Key: key, Val: val, Id: id, Counter: counter, Vec: vec_i}
	broadcast(msg)
	msg := Message{Kind: ACK, Counter: counter}
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

		msg := <-c
		
		select msg.Kind {
			case READ:
				recvRead(msg.Key, msg.Id, msg.Counter, msg.Vec)
			case WRITE:
				recvWrite(msg.Key, msg.Val, msg.Id, msg.Counter, msg.Vec)
		}
	}
}