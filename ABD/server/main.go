package main 

import(
	"sync"
	"flag"
)

var (
	mutex  = &sync.Mutex{}
	addr  string
	mData  = make(map[int]TagVal)
)

// used to mark the phase
const GET = 0
const SET = 1

func main(){
	// init storage IP
	flag.StringVar(&addr, "addr", "128.52.179.163", "input addr")
	flag.Parse()
	// create cassandra session
	server_task()
}
