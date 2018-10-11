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

func send(msg *Message, addr string){
	b := getGobFromMsg(msg)
	udpAddr,err1 := net.ResolveUDPAddr("udp4", addr)
	if err1 != nil {
		fmt.Println("Err getting addr")
	}
	// create send socket
	conn,err2 := net.DialUDP("udp", nil, udpAddr)
	if err2 != nil {
		fmt.Println("Err dialing.")
	}
	// send the data
	conn.Write(b)
	conn.Close()
}

// recv action
func (n *Node) recv(done chan bool){
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
			if msg.Type == SERVER {
				// push the message to inQueue
				heap.Push(&n.inQueue, &msg)
			} else if msg.Type == CLIENT_WRITE {
				n.write(msg.Key, msg.Val)
				rep := Message{Type: CLIENT_WRITE, Id: -1, Key: 0, Val: "", Vec: make([]int,1)}
				send(&rep, CLIENT_ADDR)
			} else if msg.Type == CLIENT_READ{
				rep := Message{Type: CLIENT_READ, Id: -1, Key: msg.Key, Val: n.read(msg.Key), Vec: make([]int,1)}
				send(&rep, CLIENT_ADDR)
			} else{
				status = false
				done<-true
				break
			}
			
		}
	}
}

func listener(){
	// resolve for udp address by membership list and id
	udpAddr,err1 := net.ResolveUDPAddr("udp4", CLIENT_ADDR)
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
			if msg.Type == CLIENT_WRITE {
				write_chan <- true
			} else {
				read_chan <- msg.Val
			}
		}
	}
}