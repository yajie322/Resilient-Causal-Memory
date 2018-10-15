package main

import (
	"bufio"
	"fmt"
	"os"
	"flag"
	"strconv"
	"strings"
)

const SERVER = 0
const CLIENT_WRITE = 1
const CLIENT_READ = 2
const QUIT = 3

var (
	id int
	// mutex = new(sync.mutex)
	svr_list = make(map[int]string)
	clt_list = make(map[int]string)
	status   bool
	node_type string
	write_chan = make(chan bool)
	read_chan = make(chan string)
)

func main() {

	flag.StringVar(&node_type, "type", "s", "specify the node type")
	flag.IntVar(&id, "id", 0, "specify the node id")
	flag.Parse()

	// read config file and get mem_list
	config, err := os.Open("config.txt")
	if err != nil {
		fmt.Print(err)
		return
	}
	scanner := bufio.NewScanner(config)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		id, err := strconv.Atoi(line[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		addr := line[1]
		if id < 3{
			svr_list[id] = addr
		} else {
			clt_list[id] = addr
		}
		
	}
	config.Close()

	status = true
	// id, _ = strconv.Atoi(os.Getenv("id"))

	if node_type == "s" {
		// // up and running
		// status = true

		var n Node

		// start running
		n.init(len(svr_list))

		done := make(chan bool)

		go n.recv(done)
		go n.send()
		go n.apply()

		<-done

	} else {
		go listener()
		// userInput()
		workload(10000)
	}
}
