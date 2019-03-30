package main 

import (
	zmq "github.com/pebbe/zmq4"
)

// init dealer socket
func createDealerSocket() *zmq.Socket {
	dealer,_ := zmq.NewSocket(zmq.DEALER)
	var addr string
	for _,server := range servers {
		addr = "tcp://" + server
		dealer.Connect(addr)
	}
	return dealer
}

func sendStore(msg Message, dealer *zmq.Socket) {
	msgToSend := getGobFromMsg(msg)
	for i := 0 ; i < WRITE_QUORUM; i++ {
		dealer.SendBytes(msgToSend.Bytes(), 0)
	}
}

// broadcast to all server
func broadcast(msg Message, dealer *zmq.Socket) {
	msgToSend := getGobFromMsg(msg)
	for i := 0 ; i < len(servers); i++ {
		dealer.SendBytes(msgToSend.Bytes(), 0)
	}
}

// recv msg with data
func recvData(dealer *zmq.Socket) TagVal {
	msgBytes,_ := dealer.RecvBytes(0)
	msg := getMsgFromGob(msgBytes)
	if msg.OpType != GET {
		return recvData(dealer)
	}
	return msg.Tv
}

// recv msg without data
func recvTs(dealer *zmq.Socket) int{
	msgBytes,_ := dealer.RecvBytes(0)
	msg := getMsgFromGob(msgBytes)
	if msg.OpType != GETTS {
		return recvTs(dealer)
	}
	return msg.Tv.Ts
}
