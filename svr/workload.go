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
func initWrite(num int){
	// write data in the form (string,blob) into table tmp
	for i:= 0; i < num; i++{
		write(i, strconv.Itoa(i))
	}
}

// write info into table
func write_load(num int, wTime chan time.Duration){
	// write data in the form (int, string) into table tmp
	// mutex.Lock()
	start := time.Now()
	write(num, strconv.Itoa(num))
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	wTime <- elapsed
}

// read info from table by key
func read_load(num int, rTime chan time.Duration){
	// write data in the form table tmp with key = num	
	// mutex.Lock()
	start := time.Now()
	read(num)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	rTime <- elapsed
}

func workload(num int){
	
	wTime := make(chan time.Duration)
	rTime := make(chan time.Duration)
	var WTotal, RTotal int = 0, 0

	// insert value into table before start testing
	for i := 0; i < num; i++{
		initWrite(i)
	}

	// go routine for concurrent read & write
	for i := 0; i < num; i++{
		go write_load(i, wTime)
		go read_load(i, rTime)
	}

	// retrieve elapsed time
	for i := 0; i < num; i++{
		WTotal += int(<-wTime/time.Millisecond)
		RTotal += int(<-rTime/time.Millisecond)
	}

	fmt.Printf("Avg write time: %f ms\n", float64(WTotal)/float64(num))
	fmt.Printf("Avg read time: %f ms\n", float64(RTotal)/float64(num))
}
