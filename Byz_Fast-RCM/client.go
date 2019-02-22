package main

import (
	"net"
	"fmt"
	"sync"
	// "strconv"
)

type ReadBufEntry struct {
	counter 	int
	val 		string
	// vec_clock	[]int
}

type Client struct {
	vec_clock		[]int
	counter			int
	readBuf 		map[ReadBufEntry]int
	readBuf_lock	sync.Mutex
	writeBuf		map[int][]int
	writeBuf_lock	sync.RWMutex
	writeBuf_cond	*sync.Cond
	localBuf		map[int]string
	val_chan		chan string
}

func (clt *Client) init(group_size int) {
	// init vector timestamp with length group_size
	clt.vec_clock = make([]int, group_size)
	for i:= 0; i < group_size; i++ {
		clt.vec_clock[i] = 0
	}
	clt.counter = 0
	// init read buffer as counter(int) - (value, timestamp) tuple (ReadBufEntry) pairs
	clt.readBuf = make(map[ReadBufEntry]int)
	clt.readBuf_lock = sync.Mutex{}
	clt.writeBuf = make(map[int] []int)
	clt.writeBuf_lock = sync.RWMutex{}
	clt.writeBuf_cond = sync.NewCond(&clt.writeBuf_lock)
	clt.localBuf = make(map[int] string)

	clt.val_chan = make(chan string)
}

func (clt *Client) read(key int) string {
	msg := Message{Kind: READ, Key: key, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	val := <-clt.val_chan
	clt.counter += 1
	return val
}

func (clt *Client) write(key int, value string) {
	clt.vec_clock[id] += 1
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	broadcast(&msg)
	clt.writeBuf_cond.L.Lock()
	entry, isIn := clt.writeBuf[clt.counter]
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

func (clt *Client) recvRESP(counter int, key int, val string, vec []int) {
	entry := ReadBufEntry{counter: counter, val: val}

	if smallerEqualExceptI(vec, clt.vec_clock, 999999) {
		val, isIn := clt.localBuf[key]
		if !isIn {
			panic("value is not in local buffer")
		}
		clt.val_chan <- val
		return
	} else {
		clt.readBuf_lock.Lock() 
		if _, isIn := clt.readBuf[entry]; isIn {
			clt.readBuf[entry] += 1
		} else {
			clt.readBuf[entry] = 1
		}
		buf_num := clt.readBuf[entry]
		clt.readBuf_lock.Unlock()

		clt.merge_clock(vec)

		if buf_num == F+1 {
			clt.val_chan <- entry.val
		}
	}

}
func (clt *Client) recvACK(counter int, vec []int) {
	// if _, isIn := clt.writeBuf[counter]; isIn {
	// 	panic("write operation " + strconv.Itoa(counter) + " is already in the write buffer")
	// }
	clt.writeBuf_lock.Lock()
	clt.writeBuf[counter] = vec
	fmt.Println("ACK message received vec", vec)
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
				clt.recvRESP(msg.Counter, msg.Key, msg.Val, msg.Vec)
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