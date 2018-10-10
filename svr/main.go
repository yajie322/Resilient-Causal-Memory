package main

import (
	"os"
	"fmt"
	"bufio"
	"strings"
    "strconv"
)

const SERVER = 0
const CLIENT_WRITE = 1
const CLIENT_READ = 2
const CLIENT_ADDR = "127.0.0.1:8080"

var(
	id int
    // mutex = new(sync.mutex)
	mem_list = make(map[int]string)
    status bool
)

func main(){
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
        mem_list[id] = addr
    }
    config.Close()

    if os.Args[1] == "s" {
        // // up and running
        // status = true
        
        var n Node
        // get node id
        id,_ = strconv.Atoi(os.Getenv("id"))

        // start running
        n.init(len(mem_list))
        
        go n.recv()
        go n.send()
        go n.apply()

    } else {
        status = true
        write_chan := make(chan bool)
        read_chan := make(chan string)
        go listener(write_chan, read_chan)
        userInput(write_chan, read_chan)
    }
}