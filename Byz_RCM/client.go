package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"time"
)

type ReadBufEntry struct {
	counter 	int
	val      	string
	// vec_clock 	[]int
}

type Client struct {
	vec_clock      	[]int
	counter        	int
	writer_ts      	map[int]map[int][]int
	readBuf        	map[ReadBufEntry]int
	val_chan	   	chan string
}

func (clt *Client) init() {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, NUM_CLIENT)
	// set vector timestamp to zero
	for i := 0; i < NUM_CLIENT; i++ {
		clt.vec_clock[i] = 0
	}
	clt.counter = 0
	// init writer_ts as counter(int) - timestamp([]int) pairs
	clt.writer_ts = make(map[int]map[int][]int)
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[ReadBufEntry]int)

	clt.val_chan = make(chan string, 1)
}

func (clt *Client) read(key int) string {
	var val string
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: READ, Key: key, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg, dealer)
	fmt.Printf("Client %d broadcasted msg READ\n", node_id)
	fmt.Println(time.Now())

	EnoughRESP: 
		for i:=0; i < len(server_list); i++ {
			clt.recvRESP(dealer)
			select {
			case val = <-clt.val_chan:
				break EnoughRESP
			default:
			}
		}

	clt.counter += 1
	return val
}

func (clt *Client) write(key int, value string) {
	var numAck int
	dealer := createDealerSocket()
	defer dealer.Close()
	clt.vec_clock[node_id] += 1
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg, dealer)
	fmt.Printf("Client %d broadcasted msg WRITE\n", node_id)
	fmt.Println(time.Now())

	EnoughACK:
		for i:=0; i < len(server_list); i++{
			clt.recvACK(dealer)
			numAck = len(clt.writer_ts[clt.counter])
			if numAck > F {
				break EnoughACK
			}
		}
	vec_set := clt.writer_ts[clt.counter]
	// merge all elements of writer_ts[counter] with local vector clock
	for _, vec := range vec_set {
		clt.merge_clock(vec)
	}
	clt.counter += 1
}

// Actions to take if receive RESP message
func (clt *Client) recvRESP(dealer *zmq.Socket) {
	msgBytes, err := dealer.RecvBytes(0)
	if err != nil {
		fmt.Println("Error occurred when client receiving ACK, err msg: ", err)
		fmt.Println(dealer.String())
	}
	msg := getMsgFromGob(msgBytes)

	if msg.Kind != RESP || msg.Counter != clt.counter {
		clt.recvRESP(dealer)
	} else {
		entry := ReadBufEntry{counter: msg.Counter, val: msg.Val}
		if _, isIn := clt.readBuf[entry]; isIn {
			clt.readBuf[entry] += 1
		} else {
			clt.readBuf[entry] = 1
		}

		clt.merge_clock(msg.Vec)

		if clt.readBuf[entry] == F+1 {
			clt.val_chan <- msg.Val
		}
	}
	
}

// Actions to take if receive ACK message
func (clt *Client) recvACK(dealer *zmq.Socket) {
	msgBytes, err := dealer.RecvBytes(0)
	if err != nil {
		fmt.Println("Error occurred when client receiving ACK, err msg: ", err)
		fmt.Println(dealer.String())
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != ACK || msg.Counter != clt.counter {
		clt.recvACK(dealer)
	} else {
		if _, isIn := clt.writer_ts[msg.Counter]; !isIn {
			clt.writer_ts[msg.Counter] = make(map[int][]int)
		}
		fmt.Println("ACK message received vec", msg.Vec, time.Now())
		clt.writer_ts[msg.Counter][len(clt.writer_ts[msg.Counter])] = msg.Vec
	}
}

// helper function that merges a vector clock with client's own vector clock
func (clt *Client) merge_clock(vec []int) {
	if len(clt.vec_clock) != len(vec) {
		fmt.Println(clt.vec_clock)
		fmt.Println(vec)
		panic("vector clocks are of different lengths")
	}
	for i := 0; i < len(clt.vec_clock); i++ {
		if vec[i] > clt.vec_clock[i] {
			clt.vec_clock[i] = vec[i]
		}
	}
}
