package main

import (
	"net"
	"fmt"
	"sync"
	// "strconv"
)

type ReadBufEntry struct {
	val 		string
	vec_clock	[]int
}

type Client struct {
	vec_clock		[]int
	counter			int
	readBuf 		map[int]ReadBufEntry
	readBuf_lock	sync.RWMutex
	readBuf_cond	*sync.Cond
	writeBuf		map[int][]int
	writeBuf_lock	sync.RWMutex
	writeBuf_cond	*sync.Cond
	localBuf		map[int]string
}

func (clt *Client) init(group_size int) {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, group_size)
	for i:= 0; i < group_size; i++ {
		clt.vec_clock[i] = 0
	}
	clt.counter = 0
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[int] ReadBufEntry)
	clt.readBuf_lock = sync.RWMutex{}
	clt.readBuf_cond = sync.NewCond(&clt.readBuf_lock)
	clt.writeBuf = make(map[int] []int)
	clt.writeBuf_lock = sync.RWMutex{}
	clt.writeBuf_cond = sync.NewCond(&clt.writeBuf_lock)
	clt.localBuf = make(map[int] string)
}

func (clt *Client) read(key int) string {
	msg := Message{Kind: READ, Key: key, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	clt.readBuf_lock.RLock()
	entry, isIn := clt.readBuf[clt.counter]
	clt.readBuf_lock.RUnlock()
	clt.readBuf_cond.L.Lock()
	for !isIn {
		clt.readBuf_cond.Wait()
		entry, isIn = clt.readBuf[clt.counter]
	}
	clt.readBuf_cond.L.Unlock()
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
	clt.vec_clock[id] += 1
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	clt.writeBuf_lock.RLock()
	entry, isIn := clt.writeBuf[clt.counter]
	clt.writeBuf_lock.RUnlock()
	clt.writeBuf_cond.L.Lock()
	for !isIn {
		clt.writeBuf_cond.Wait()
		entry, isIn = clt.writeBuf[clt.counter]
	}
	clt.writeBuf_cond.L.Unlock()
	clt.merge_clock(entry)
	clt.localBuf[key] = value
	clt.counter += 1
	// return WRITE-ACK
}

func (clt *Client) recvRESP(counter int, val string, vec []int) {
	entry := ReadBufEntry{val: val, vec_clock: vec}
	clt.readBuf_lock.Lock()
	clt.readBuf[counter] = entry
	clt.readBuf_cond.Broadcast()
	clt.readBuf_lock.Unlock()
}
func (clt *Client) recvACK(counter int, vec []int) {
	// if _, isIn := clt.writeBuf[counter]; isIn {
	// 	panic("write operation " + strconv.Itoa(counter) + " is already in the write buffer")
	// }
	clt.writeBuf_lock.Lock()
	clt.writeBuf[counter] = vec
	clt.writeBuf_cond.Broadcast()
	clt.writeBuf_lock.Unlock()
}

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