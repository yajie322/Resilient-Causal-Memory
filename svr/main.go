package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"strings"
    "strconv"
)

var(
	id int
	mem_list = make(map[int]string)
)

func main(){
	var n Node

	// get node id
	flag.IntVar(&id, "id", 0, "specify the node id")

	// read config file
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
        mem_list[id] = addr
    }
    config.Close()

    // start running
	n.init(len(mem_list))
	go n.send()
	go n.recv()
}