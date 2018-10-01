package main 

import (
	"net"
	"fmt"
	"container/heap"
)

// use udp to broadcast msg
func broadcast(msg *Message){
	//use gob to serialized data before sending
	b := getGobFromMsg(msg)
	for Id,addr := range mem_list {
		if (Id != id) {
			// resolve address
			udpAddr,err1 := net.ResolveUDPAddr("udp4", addr)
			if err1 != nil {
				fmt.Println("Err getting addr")
				return //gracefully deal with connection error
			}
			// create send socket
			conn,err2 := net.DialUDP("udp", nil, udpAddr)
			if err2 != nil {
				fmt.Println("Err dialing.")
				return
			}
			// send the data
			conn.Write(b)
			conn.Close()
		}
	}
}

// recv action
func (n *Node) recv(){
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
}