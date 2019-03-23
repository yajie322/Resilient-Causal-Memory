package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
)

func creatDealerSocket(target string) *zmq.Socket {
	dealer,_ := zmq.NewSocket(zmq.DEALER)
	dealer.Connect("tcp://" + target)
	return dealer
}

func createPublisherSocket(pubAddr string) *zmq.Socket {
	publisher,_ := zmq.NewSocket(zmq.PUB)
	publisher.Bind("tcp://" + pubAddr)
	return publisher
}


func createSubscriberSocket() *zmq.Socket {
	subscriber,_ := zmq.NewSocket(zmq.SUB)
	var addr string
	for _,server := range svr_list {
		addr = "tcp://" + server
		subscriber.Connect(addr)
	}
	subscriber.SetSubscribe(FILTER)
	return subscriber
}

func (svr *Server)publish(msg *Message) {
	b := getGobFromMsg(msg)
	if _, err := svr.publisher.SendBytes(b,0); err != nil {
		fmt.Println("Error occurred at line 42 in file zmqConn.go", err)
	}
}

func send(msg *Message, soc *zmq.Socket){
	b := getGobFromMsg(msg)
	if _,err := soc.SendBytes(b,0); err != nil{
		fmt.Println("Err when sending: {}",err)
	}
}

func recvACK(soc *zmq.Socket){
	msgBytes, err := soc.RecvBytes(0)
	if err != nil {
		fmt.Println("Error occurred when client receiving ACK, err msg: ", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Type != ACK {
		fmt.Print("wrong msg type")
		recvACK(soc)
	}
	return
}

func recvData(soc *zmq.Socket) string{
	msgBytes, err := soc.RecvBytes(0)
	if err != nil {
		fmt.Println("Error occurred when client receiving Data, err msg: ", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Type != DATA {
		fmt.Print("wrong msg type")
		recvData(soc)
	}
	return msg.Val
}

func (svr *Server) server(svrAddr string) {
	// Set the ZMQ sockets
	frontend,_ := zmq.NewSocket(zmq.ROUTER)
	defer fmt.Println("frontend socket closed")
	defer frontend.Close()
	frontend.Bind("tcp://" + svrAddr)

	//  Backend socket talks to workers over inproc
	backend, _ := zmq.NewSocket(zmq.DEALER)
	defer fmt.Println("backend socket closed")
	defer backend.Close()
	backend.Bind("inproc://backend")

	go svr.serverWorker()

	//  Connect backend to frontend via a proxy
	err := zmq.Proxy(frontend, backend, nil)
	log.Fatal("Proxy interrupted:", err)
}

func (svr *Server)serverWorker() {
	worker, _ := zmq.NewSocket(zmq.DEALER)
	defer fmt.Println("worker socket closed")
	defer worker.Close()
	worker.Connect("inproc://backend")
	msgReply := make([][]byte, 2)

	for i := 0; i < len(msgReply); i++ {
		// the first frame  specifies the identity of the sender, the second specifies the content
		msgReply[i] = make([]byte, 0)
	}

	for {
		msg,err := worker.RecvMessageBytes(0)
		if err != nil {
			fmt.Println(err)
		}
		// decode message
		message := getMsgFromGob(msg[1])
		if message.Type == ERROR{
			continue
		}
		msgReply[0] = msg[0]

		// create response message
		tmpMsg := svr.createRep(message)
		// encode message
		tmpGob := getGobFromMsg(tmpMsg)
		msgReply[1] = tmpGob

		numBytes, err := worker.SendMessage(msgReply)
		if err != nil {
			fmt.Println("Error occurred when server worker sending reply, # of bytes sent is ", numBytes)
		}
	}
}