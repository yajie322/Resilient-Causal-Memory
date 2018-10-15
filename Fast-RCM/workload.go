package main 

import(
	"strconv"
	"fmt"
	"time"
	// "sync"
	"math/rand"
	"sort"
	"strings"
)

// initialize mutex lock
// var mutex = &sync.Mutex{}
const READ_PORTION = 0.3
const DATA_SIZE = 64

// write info into table before read, with out tracking time
func (clt *Client) initWrite(num int){
	// write data in the form (string,blob) into table tmp
	for i:= 0; i < num; i++{
		clt.write(i, strings.Repeat(strconv.Itoa(i % 10), DATA_SIZE))
	}
}

// write info into table
func (clt *Client) write_load(num int, val string, wTime chan time.Duration){
	// write data in the form (int, string) into table tmp
	// mutex.Lock()
	start := time.Now()
	clt.write(num, val)
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
	num_read := 0
	num_write := 0

	wTime := make(chan time.Duration)
	rTime := make(chan time.Duration)
	var WTotal, RTotal int = 0, 0

	// insert value into table before start testing
	for i := 0; i < num_read; i++{
		clt.initWrite(i%10)
	}

	for i := 0; i < num; i++ {
		temp := rand.Float64()
		if temp < READ_PORTION {
			clt.read_load(i%10, rTime)
			num_read += 1
		} else {
			clt.write_load(i%10, strings.Repeat(strconv.Itoa(i % 10), DATA_SIZE), wTime)
			num_write += 1
		}
	}

	// // go routine for concurrent write
	// for i := 0; i < num_write; i++{
	// 	go clt.write_load(i, wTime)
	// }
	// // go routine for concurrent read
	// for i := 0; i < num_read; i++{
	// 	go clt.read_load(i, rTime)
	// }

	// retrieve elapsed time
	write_times := make([]int, num_write)
	for i := 0; i < num_write; i++{
		millisec := int(<-wTime/time.Millisecond)
		WTotal += millisec
		write_times[i] = millisec
	}
	// retrieve elapsed time
	read_times := make([]int, num_read)
	for i := 0; i < num_read; i++{
		millisec := int(<-rTime/time.Millisecond)
		RTotal += millisec
		read_times[i] = millisec
	}

	sort.Ints(write_times)
	sort.Ints(read_times)

	fmt.Printf("Avg write time: %f ms\n", float64(WTotal)/float64(num_write))
	fmt.Printf("Avg read time: %f ms\n", float64(RTotal)/float64(num_read))
	fmt.Printf("95-th percentile for write time: %d ms\n", write_times[int(float64(len(write_times)) * 0.95)])
	fmt.Printf("95-th percentile for read time: %d ms\n", read_times[int(float64(len(read_times)) * 0.95)])
}
