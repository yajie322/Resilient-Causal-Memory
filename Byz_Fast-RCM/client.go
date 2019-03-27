package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

type ReadBufKey struct {
	counter 	int
	val 		string
	// vec_clock	[]int
}

type Client struct {
	vec_clock		[]int
	counter			int
	readBuf 		map[ReadBufKey]int
	writeBuf		map[int][]int
	localBuf		map[int]string
	val_chan		chan string
}

func (clt *Client) init() {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, NUM_CLIENT)
	for i:= 0; i < NUM_CLIENT; i++ {
		clt.vec_clock[i] = 0
	}
	clt.counter = 0
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufKey) pairs
	clt.readBuf = make(map[ReadBufKey]int)
	clt.writeBuf = make(map[int] []int)
	clt.localBuf = make(map[int] string)

	clt.val_chan = make(chan string, 1)
}

func (clt *Client) read(key int) string {
	var val string
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: READ, Key: key, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg, dealer)
	fmt.Printf("Client %d broadcasted msg READ\n", node_id)

	for i := 0; i < len(server_list); i++ {
		clt.recvRESP(dealer)
		select {
		case val = <-clt.val_chan:
			break
		default:
		}
	}
	clt.counter += 1
	return val
}

func (clt *Client) write(key int, value string) {
	dealer := createDealerSocket()
	defer dealer.Close()
	clt.vec_clock[node_id] += 1
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: node_id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg, dealer)
	fmt.Printf("Client %d broadcasted msg WRITE\n", node_id)

	var vec []int
	for i := 0; i < len(server_list); i++{
		clt.recvACK(dealer)
		if _,isIn := clt.writeBuf[clt.counter]; isIn {
			break
		}
	}
	clt.merge_clock(vec)
	clt.localBuf[key] = value
	clt.counter += 1
	// return WRITE-ACK
}

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
		if smallerEqualExceptI(msg.Vec, clt.vec_clock, 999999) {
			val, isIn := clt.localBuf[msg.Key]
			if !isIn {
				panic("value is not in local buffer")
			}
			clt.val_chan <- val
		} else {
			entry := ReadBufKey{counter: msg.Counter, val: msg.Val}
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
}
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