package main

import (
	zmq "github.com/pebbe/zmq4"
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

func createPublisherSocket() *zmq.Socket {
	publisher,_ := zmq.NewSocket(zmq.PUB)
	publisher.Bind("tcp://*:"+"9999")
	var addr string
	for _,server := range mem_list {
		addr = "tcp://" + server
		publisher.Connect(addr)
	}
	return publisher
}

// broadcast to all
func zmqBroadcast(msg *Message, dealer *zmq.Socket){
	//use gob to serialized data before sending
	b := getGobFromMsg(msg)
	for i := 0; i < len(mem_list); i++ {
		dealer.SendBytes(b,0)
	}
}