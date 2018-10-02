package main

import (
	"os"
	"fmt"
	// "flag"
    // "sync"
	"bufio"
	"strings"
    "strconv"
)

var(
	id int
    // mutex = new(sync.mutex)
	mem_list = make(map[int]string)
    status bool
)

func main(){
	var n Node

    // up and running
    status = true

	// get node id
	id,_ = strconv.Atoi(os.Getenv("id"))
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
	
	go n.recv()
    go n.send()
    go n.apply()

    n.userInput()
}

func (n *Node) userInput(){
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
            n.write(key, input[2])
        } else if strings.HasPrefix(text, "read") {
            key, err := strconv.Atoi(strings.SplitN(text, " ", 2)[1])
            if err != nil {
                fmt.Println(err)
            }
            // output read
            fmt.Printf("\t%s\n",n.read(key))
        } else {
            // quit
            // mutex.Lock()
            status = false
            // mutex.Unlock()
            break
        }
    }
}