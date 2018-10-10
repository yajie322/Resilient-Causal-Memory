package main 

import (
	"net"
	"fmt"
)

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