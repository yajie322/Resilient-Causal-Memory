package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"net"
)

func createDealerSocket() *zmq.Socket {
	dealer,_ := zmq.NewSocket(zmq.DEALER)
	var addr string
	for _,server := range mem_list {
		addr = "tcp://" + server
		dealer.Connect(addr)
	}
	return dealer
}

func send(msg *Message, addr string) {
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
	conn.Write(b)
	conn.Close()

}

// broadcast to all
func zmqBroadcast(msg *Message, dealer *zmq.Socket){
	//use gob to serialized data before sending
	b := getGobFromMsg(msg)
	for i := 0; i < len(mem_list); i++ {
		dealer.SendBytes(b.Bytes(),0)
	}
}