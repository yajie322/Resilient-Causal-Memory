package main

import (
	"math/rand"
)

func write(key int, value string) {
	// chose server
	chosen := rand.Intn(len(svr_list))
	// create socket
	dealer := creatDealerSocket(svr_list[chosen])
	defer dealer.Close()
	//create message & send
	msg := Message{Type: WRITE, Id: node_id, Key: key, Val: value, Vec: make([]int, 1)}
	send(&msg, dealer)
	// get ACK
	recvACK(dealer)
	return
}

func read(key int) string {
	// chose server
	chosen := rand.Intn(len(svr_list))
	// create socket
	dealer := creatDealerSocket(svr_list[chosen])
	defer dealer.Close()
	// create message object & send
	msg := Message{Type: READ, Id: node_id, Key: key, Val: "", Vec: make([]int, 1)}
	send(&msg, dealer)
	// get result
	res := recvData(dealer)
	return res
}