package main

import (
	"bufio"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"sync"
	"os"
	"flag"
	"strconv"
	"strings"
)

var (
	node_id int
	mutex = &sync.Mutex{}
	svr_list = make(map[int]string)
	clt_list = make(map[int]string)
	status   bool
	node_type string


)

func main() {

	flag.StringVar(&node_type, "type", "s", "specify the node type")
	flag.IntVar(&node_id, "id", 0, "specify the node id")
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

		var svr Server

		// start running
		svr.init(len(svr_list))
		go svr.send()
		go svr.apply()
		svr.server(svr_list[node_id])
	} else {
		// userInput()
		workload(10000)
	}
}
