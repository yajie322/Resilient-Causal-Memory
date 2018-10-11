package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func userInput(write_chan chan bool, read_chan chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("->")
		// handle command line input
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if strings.HasPrefix(text, "write") {
			input := strings.SplitN(text, " ", 3)

			key, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println(err)
			}
			// write
			write(key, input[2], write_chan)
		} else if strings.HasPrefix(text, "read") {
			key, err := strconv.Atoi(strings.SplitN(text, " ", 2)[1])
			if err != nil {
				fmt.Println(err)
			}
			// output read
			fmt.Printf("\t%s\n", read(key, read_chan))
		} else {
			status = false
			msg := Message{Type: QUIT, Id: 0, Key: 0, Val: "", Vec: make([]int, 1)}
			broadcast(&msg)
			break
		}
	}
}

func write(key int, value string, done chan bool) {
	// create message object
	msg := Message{Type: CLIENT_WRITE, Id: 0, Key: key, Val: value, Vec: make([]int, 1)}
	// chose server
	chosen := rand.Intn(len(mem_list))
	// send
	send(&msg, mem_list[chosen])
	// get result from listener
	<-done
	return
}

func read(key int, done chan string) string {
	// create message object
	msg := Message{Type: CLIENT_READ, Id: 0, Key: key, Val: "", Vec: make([]int, 1)}
	// chose server
	chosen := rand.Intn(len(mem_list))
	// send
	send(&msg, mem_list[chosen])
	// get result from listener
	return <-done
}
