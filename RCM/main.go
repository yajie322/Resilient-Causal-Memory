package main

import (
	"os"
	"fmt"
	"flag"
    // "time"
    // "sync"
	"bufio"
	"strings"
    "strconv"
)

var(
	id int
    node_type string
    // mutex = new(sync.mutex)
	mem_list = make(map[int]string)
    status bool
)

func main(){
    //node_type = os.Getenv("type")
    flag.StringVar(&node_type, "type", "server", "specify the node type")

    // up and running
    status = true

	// get node id
	// id,_ = strconv.Atoi(os.Getenv("id"))
    flag.IntVar(&id, "id", 0, "specify the node id")
    flag.Parse()
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

    switch node_type {
    case "server":
        var node Server
        node.init(len(mem_list))
        go node.recv()
        for status {
            <- node.update_needed
            go node.update()
        }
    case "client":
        var node Client
        node.init(len(mem_list))
        go node.recv()
        // node.userInput()
        node.workload(10000)
    }
}

func (clt *Client) userInput(){
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("->")
        // handle command line input
        text,_ := reader.ReadString('\n')
        text = strings.Replace(text, "\n", "", -1)
        if strings.HasPrefix(text, "write"){
            input := strings.SplitN(text, " ", 3)
            
            key, err := strconv.Atoi(input[1])
            if err != nil {
                fmt.Println(err)
            }
            // write
            clt.write(key, input[2])
        } else if strings.HasPrefix(text, "read") {
            key, err := strconv.Atoi(strings.SplitN(text, " ", 2)[1])
            if err != nil {
                fmt.Println(err)
            }
            // output read
            fmt.Printf("\t%s\n", clt.read(key))
        } else {
            // quit
            // mutex.Lock()
            status = false
            // mutex.Unlock()
            break
        }
    }
}

