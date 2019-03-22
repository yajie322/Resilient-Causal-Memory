package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func write(key int, value string) {
	// create message object
	msg := Message{Type: CLIENT_WRITE, Id: id, Key: key, Val: value, Vec: make([]int, 1)}
	// chose server
	chosen := rand.Intn(len(svr_list))
	// send
	send(&msg, svr_list[chosen])
	// get result from listener
	<-write_chan
	return
}

func read(key int) string {
	// create message object
	msg := Message{Type: CLIENT_READ, Id: id, Key: key, Val: "", Vec: make([]int, 1)}
	// chose server
	chosen := rand.Intn(len(svr_list))
	// send
	send(&msg, svr_list[chosen])
	// get result from listener
	return <-read_chan
}
