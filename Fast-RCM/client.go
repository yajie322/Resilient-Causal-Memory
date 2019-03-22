package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
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
	//readBuf_lock	sync.RWMutex
	//readBuf_cond	*sync.Cond
	writeBuf		map[int][]int
	//writeBuf_lock	sync.RWMutex
	//writeBuf_cond	*sync.Cond
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
	//clt.readBuf_lock = sync.RWMutex{}
	//clt.readBuf_cond = sync.NewCond(&clt.readBuf_lock)
	clt.writeBuf = make(map[int] []int)
	//clt.writeBuf_lock = sync.RWMutex{}
	//clt.writeBuf_cond = sync.NewCond(&clt.writeBuf_lock)
	clt.localBuf = make(map[int] string)
}

func (clt *Client) read(key int) string {
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: READ, Key: key, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)
	for i:=0; i < len(svrList); i++{
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
	clt.vec_clock[id] += 1
	dealer := createDealerSocket()
	defer dealer.Close()
	msg := Message{Kind: WRITE, Key: key, Val: value, Id: id, Counter: clt.counter, Vec: clt.vec_clock}
	zmqBroadcast(&msg,dealer)

	for i:=0; i < len(svrList); i++{
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
		fmt.Println("client.go line 94, dealer RecvBytes: {}", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != RESP {
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
		fmt.Println("client.go line 108, dealer RecvBytes: {}", err)
	}
	msg := getMsgFromGob(msgBytes)
	if msg.Kind != ACK {
		clt.recvACK(dealer)
	} else{
		clt.writeBuf[msg.Counter] = msg.Vec
		fmt.Println("ACK message received vec", msg.Vec)
	}
}

//func (clt *Client) recv(){
//	// resolve for udp address by membership list and id
//	udpAddr,err1 := net.ResolveUDPAddr("udp4", svrList[id])
//	if err1 != nil {
//		fmt.Println("address not found")
//	}
//
//	// create listner socket by address
//	conn,err2 := net.ListenUDP("udp", udpAddr)
//	if err2 != nil {
//		fmt.Println("address can't listen")
//	}
//	defer conn.Close()
//
//	for status {
//		c := make(chan Message)
//
//		go func() {
//			//buffer size is 1024 bytes
//			buf := make([]byte, 1024)
//			num,_,err3 := conn.ReadFromUDP(buf)
//			if err3 != nil {
//				fmt.Println(err3)
//			}
//			//deserialize the received data and output to channel
//			c <- getMsgFromGob(buf[:num])
//		}()
//
//		msg := <-c
//
//		switch msg.Kind {
//			case RESP:
//				clt.recvRESP(msg.Counter, msg.Val, msg.Vec)
//			case ACK:
//				clt.recvACK(msg.Counter, msg.Vec)
//		}
//	}
//}

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