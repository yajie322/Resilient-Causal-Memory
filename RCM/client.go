package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"sync"
)

type ReadBufEntry struct {
	val       string
	vec_clock []int
}

type Client struct {
	vec_clock      []int
	counter        int
	writer_ts      map[int]map[int][]int
	writer_ts_lock sync.Mutex
	writer_ts_cond *sync.Cond
	readBuf        map[int]ReadBufEntry
	readBuf_lock   sync.Mutex
	readBuf_cond   *sync.Cond
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
	clt.writer_ts_lock = sync.Mutex{}
	clt.writer_ts_cond = sync.NewCond(&clt.writer_ts_lock)
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[int]ReadBufEntry)
	clt.readBuf_lock = sync.Mutex{}
	clt.readBuf_cond = sync.NewCond(&clt.readBuf_lock)
}

func (clt *Client) read(key int) string {
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: READ, Key: key, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)
	fmt.Printf("Client %d broadcasted msg READ\n", id)

	for i:=0; i < len(mem_list); i++{
		clt.recvACK(dealer)
		clt.readBuf_lock.Lock()
		_,isIn := clt.readBuf[clt.counter]
		clt.readBuf_lock.Unlock()
		if isIn{
			break
		}
	}
	entry,_ := clt.readBuf[clt.counter]
	clt.merge_clock(entry.vec_clock)
	clt.counter += 1
	return entry.val
}

func (clt *Client) write(key int, value string) {
	var numAck int
	dealer := createDealerSocket()
	defer dealer.Close()
	clt.vec_clock[id] += 1
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)
	fmt.Printf("Client %d broadcasted msg WRITE\n", id)

	for i:=0; i < len(mem_list); i++{
		clt.recvACK(dealer)
		clt.writer_ts_lock.Lock()
		numAck = len(clt.writer_ts[clt.counter])
		clt.writer_ts_lock.Unlock()
		if numAck > F{
			break
		}
	}
	clt.writer_ts_lock.Lock()
	vec_set := clt.writer_ts[clt.counter]
	clt.writer_ts_lock.Unlock()
	// merge all elements of writer_ts[counter] with local vector clock
	for _, vec := range vec_set {
		clt.merge_clock(vec)
	}
	clt.counter += 1
}

// Actions to take if receive RESP message
func (clt *Client) recvRESP(dealer *zmq.Socket) {
	msgBytes,_ := dealer.RecvBytes(0)
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != RESP {
		clt.recvACK(dealer)
	} else{
		entry := ReadBufEntry{val: msg.Val, vec_clock: msg.Vec}
		clt.readBuf_lock.Lock()
		clt.readBuf[msg.Counter] = entry
		clt.readBuf_cond.Broadcast()
		clt.readBuf_lock.Unlock()
	}
	return
}

// Actions to take if receive ACK message
func (clt *Client) recvACK(dealer *zmq.Socket) {
	fmt.Println("recv ACK")
	msgBytes,_ := dealer.RecvBytes(0)
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != ACK {
		clt.recvRESP(dealer)
	} else{
		clt.writer_ts_lock.Lock()
		if _, isIn := clt.writer_ts[msg.Counter]; !isIn {
			clt.writer_ts[msg.Counter] = make(map[int][]int)
		}
		fmt.Println("ACK message received vec", msg.Vec)
		clt.writer_ts[msg.Counter][len(clt.writer_ts[msg.Counter])] = msg.Vec
		clt.writer_ts_cond.Broadcast()
		clt.writer_ts_lock.Unlock()
	}
	return
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
