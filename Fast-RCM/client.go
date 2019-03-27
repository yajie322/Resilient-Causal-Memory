package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

type ReadBufEntry struct {
	val 		string
	vec_clock	[]int
}

type Client struct {
	vec_clock		[]int
	counter			int
	readBuf 		map[int]ReadBufEntry
	writeBuf		map[int][]int
	localBuf		map[int]string
}

func (clt *Client) init() {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, NUM_CLIENT)
	for i:= 0; i < NUM_CLIENT; i++ {
		clt.vec_clock[i] = 0
	}
	clt.counter = 0
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[int] ReadBufEntry)
	clt.writeBuf = make(map[int] []int)
	clt.localBuf = make(map[int] string)
}

func (clt *Client) read(key int) string {
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: READ, Key: key, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)
	fmt.Printf("Client %d broadcasted msg READ\n", node_id)

	for i := 0; i < len(server_list); i++{
		clt.recvRESP(dealer)
		if _,isIn := clt.readBuf[clt.counter]; isIn{
			break
		}
	}
	entry,_ := clt.readBuf[clt.counter]
	clt.counter += 1
	if smallerEqualExceptI(entry.vec_clock, clt.vec_clock, 999999) {
		val, isIn := clt.localBuf[key]
		if !isIn {
			panic("value is not in local buffer")
			return "Error"
		}
		return val
	} else {
		clt.merge_clock(entry.vec_clock)
		return entry.val
	}
}

func (clt *Client) write(key int, value string) {
	clt.vec_clock[node_id] += 1
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)

	for i := 0; i < len(server_list); i++{
		clt.recvACK(dealer)
		if _,isIn := clt.writeBuf[clt.counter]; isIn{
			break
		}
	}
	entry,_ := clt.writeBuf[clt.counter]
	clt.merge_clock(entry)
	clt.localBuf[key] = value
	clt.counter += 1
}

func (clt *Client) recvRESP(dealer *zmq.Socket){
	msgBytes,err := dealer.RecvBytes(0)
	if err != nil{
		fmt.Println("client.go line 84, dealer RecvBytes: {}", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != RESP || msg.Counter != clt.counter {
		clt.recvRESP(dealer)
	} else{
		clt.readBuf[msg.Counter] = ReadBufEntry{val: msg.Val, vec_clock: msg.Vec}
		fmt.Println("RESP message received vec", msg.Vec)
	}
}

func (clt *Client) recvACK(dealer *zmq.Socket) {
	// counter int, vec []int
	msgBytes,err := dealer.RecvBytes(0)
	if err != nil{
		fmt.Println("client.go line 99, dealer RecvBytes: {}", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != ACK || msg.Counter != clt.counter {
		clt.recvACK(dealer)
	} else{
		clt.writeBuf[msg.Counter] = msg.Vec
		fmt.Println("ACK message received vec", msg.Vec)
	}
}

func (clt *Client) merge_clock(vec []int) {
	if len(clt.vec_clock) != len(vec) {
		panic("vector clocks are of different lengths")
	}
	for i:=0; i<len(clt.vec_clock); i++ {
		if vec[i] > clt.vec_clock[i] {
			clt.vec_clock[i] = vec[i]
		}
	}
}