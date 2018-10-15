package main 

import(
	"strconv"
	"fmt"
	"time"
	"sync"
)

// initialize mutex lock
var mutex = &sync.Mutex{}

// write info into table before read, with out tracking time
func (clt *Client) initWrite(num int){
	// write data in the form (string,blob) into table tmp
	for i:= 0; i < num; i++{
		clt.write(i, strconv.Itoa(i))
	}
}

// write info into table
func (clt *Client) write_load(num int, wTime chan time.Duration){
	// write data in the form (int, string) into table tmp
	// mutex.Lock()
	start := time.Now()
	clt.write(num, strconv.Itoa(num))
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	wTime <- elapsed
}

// read info from table by key
func (clt *Client) read_load(num int, rTime chan time.Duration){
	// write data in the form table tmp with key = num	
	// mutex.Lock()
	start := time.Now()
	clt.read(num)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	rTime <- elapsed
}

func (clt *Client) workload(num int){
	num_read = num*0.7
	num_write = num*0.3
	wTime := make(chan time.Duration)
	rTime := make(chan time.Duration)
	var WTotal, RTotal int = 0, 0

	// insert value into table before start testing
	for i := 0; i < num_read; i++{
		clt.initWrite(i)
	}

	// go routine for concurrent write
	for i := 0; i < num_write; i++{
		go clt.write_load(i, wTime)
	}
	// go routine for concurrent read
	for i := 0; i < num_read; i++{
		go clt.read_load(i, rTime)
	}

	// retrieve elapsed time
	for i := 0; i < num_write; i++{
		WTotal += int(<-wTime/time.Millisecond)
	}
	// retrieve elapsed time
	for i := 0; i < num_read; i++{
		RTotal += int(<-rTime/time.Millisecond)
	}

	fmt.Printf("Avg write time: %f ms\n", float64(WTotal)/float64(num_write))
	fmt.Printf("Avg read time: %f ms\n", float64(RTotal)/float64(num_read))
}
