package main

import (
 	zmq "github.com/pebbe/zmq4"
)

func createDealerSocket() *zmq.Socket {
	dealer,_ := zmq.NewSocket(zmq.DEALER)
	var addr string
	for _,server := range server_list {
		addr = "tcp://" + server
		dealer.Connect(addr)
	}
	return dealer
}

func createPublisherSocket(port string) *zmq.Socket {
	publisher,_ := zmq.NewSocket(zmq.PUB)
	publisher.Bind("tcp://*:" + port)
	return publisher
}


func createSubscriberSocket() *zmq.Socket {
	subscriber,_ := zmq.NewSocket(zmq.SUB)
	var addr string
	for key,server := range server_pub {
		if key != id {
			addr = "tcp://" + server
			subscriber.Connect(addr)
		}
	}
	subscriber.SetSubscribe(FILTER)
	return subscriber
}

func publish(msg *Message, publisher *zmq.Socket) {
	b := getGobFromMsg(msg)
	fmt.Println(b)
	for i := 0; i < len(server_list)-1; i++ {
		//publisher.Send(FILTER, zmq.SNDMORE)
		publisher.SendBytes(b,0)
	}
}

// broadcast to all
func zmqBroadcast(msg *Message, dealer *zmq.Socket){
	//use gob to serialized data before sending
	b := getGobFromMsg(msg)
	for i := 0; i < len(server_list); i++ {
		dealer.SendBytes(b,0)
	}
}