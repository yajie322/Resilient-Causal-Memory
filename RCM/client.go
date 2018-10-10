package main

import (
	"time"
	"net"
	"fmt"
)

type ReadBufEntry struct {
	val 		string
	vec_clock	[]int
}

type Client struct {
	vec_clock	[]int
	counter		int
	writer_ts	map[int][]int
	readBuf 	map[int]ReadBufEntry
}

func (clt *Client) init(group_size int) {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, group_size)
	clt.counter = 0
	// init writer_ts as counter(int) - timestamp([]int) pairs
	clt.writer_ts = make(map[int] []int)
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[int] ReadBufEntry)

}

func (clt *Client) read(key int) string {
	msg := Message{Kind: READ, Key: key, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	entry, isIn := clt.readBuf[clt.counter]
	for !isIn {
		time.Sleep(time.Millisecond)
		entry, isIn := clt.readBuf[clt.counter]
	}
	clt.merge_clock(entry.vec_clock)
	clt.counter += 1
	return entry.val
}

func (clt *Client) write(key int, value string) {
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	for clt.writer_ts[clt.counter] <= F {
		time.Sleep(time.Millisecond)
	}
	clt.merge_clock(clt.writer_ts[clt.counter])
	clt.counter += 1
	// return WRITE-ACK
}

// Actions to take if receive RESP message
func (clt *Client) recvRESP(counter int, val string, vec []int) {
	entry := ReadBufEntry{val: val, vec_clock: vec}
	if _, isIn := clt.readBuf[counter]; isIn {
		// panic()
	}
	clt.readBuf[counter] = entry
}

// Actions to take if receive ACK message
func (clt *Client) recvACK(counter int, vec []int) {
	if _, isIn := clt.writer_ts[counter]; isIn {
		// panic()
	}
	clt.writer_ts[counter] = vec
}

// Client listener
func (clt *Client) recv(){
	// resolve for udp address by membership list and id
	udpAddr,err1 := net.ResolveUDPAddr("udp4", mem_list[id])
	if err1 != nil {
		fmt.Println("address not found")
	}

	// create listner socket by address
	conn,err2 := net.ListenUDP("udp", udpAddr)
	if err2 != nil {
		fmt.Println("address can't listen")
	}
	defer conn.Close()

	for status {
		c := make(chan Message)

		go func() {
			//buffer size is 1024 bytes
			buf := make([]byte, 1024)
			num,_,err3 := conn.ReadFromUDP(buf)
			if err3 != nil {
				fmt.Println(err3)
			}
			//deserialize the received data and output to channel
			c <- getMsgFromGob(buf[:num])
		}()

		msg := <-c

		switch msg.Kind {
		case RESP:
			clt.recvRESP(msg.Counter, msg.Val, msg.Vec)
		case ACK:
			clt.recvACK(msg.Counter, msg.Vec)
		}
	}
}

// helper function that merges a vector clock with client's own vector clock
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