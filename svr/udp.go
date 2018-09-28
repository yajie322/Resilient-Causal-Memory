package main 

import (
	"net"
	"fmt"
	"log"
)

// use udp to broadcast msg
func (n *Node) broadcast(msg Message){
	for _,addr := range n.mem_list {
		if (entry.Id != n.id) {
			//use gob to serialized data before sending
			b := getGobFromMsg(msg)
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
			conn.Write([]byte(b))
			conn.Close()
		}
	}
}

// To Do
func (n *Node) send(){
	for n.outQueue.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		msg := item.value
		n.broadcast(msg)
	}
}

// creates listener
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

	for {
		var msg Message
		//buffer size is 1024 bytes
		buf := make([]byte, 1024)
		num,_,err3 := conn.ReadFromUDP(buf)
		if err3 != nil {
			log.Fatal(err3)
		}

		//deserialize the received data
		msg := getMsgFromGob(buf[:num])
		// push the message to inQueue
		item := &Item{value: msg}
		heap.Push(&n.inQueue, item)
	}
}