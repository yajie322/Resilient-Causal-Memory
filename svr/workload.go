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
func initWrite(num int){
	// write data in the form (string,blob) into table tmp
	for i:= 0; i < num; i++{
		write(i, strings.Repeat(strconv.Itoa(i % 10), DATA_SIZE))
	}
}

// write info into table
func write_load(num int, val string) time.Duration {
	// write data in the form (int, string) into table tmp
	// mutex.Lock()
	start := time.Now()
	write(num, val)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	return elapsed
}

// read info from table by key
func read_load(num int) time.Duration {
	// write data in the form table tmp with key = num	
	// mutex.Lock()
	start := time.Now()
	read(num)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	// send elapsed time to main thread
	return elapsed
}

func workload(num int){
	var write_times []int
	var read_times []int

	num_read := 0
	num_write := 0

	// wTime := make(chan time.Duration)
	// rTime := make(chan time.Duration)
	var WTotal, RTotal int = 0, 0

	// insert value into table before start testing
	initWrite(10)

	for i := 0; i < num; i++ {
		temp := rand.Float64()
		if temp < READ_PORTION {
			millisec := int(read_load(i%10)/time.Millisecond)
			RTotal += millisec
			read_times = append(read_times, millisec)
			num_read += 1
		} else {
			millisec := int(write_load(i%10, strings.Repeat(strconv.Itoa(i % 10), DATA_SIZE))/time.Millisecond)
			WTotal += millisec
			write_times = append(write_times, millisec)
			num_write += 1
		}
	}

	// // go routine for concurrent write
	// for i := 0; i < num_write; i++{
	// 	go write_load(i, wTime)
	// }
	// // go routine for concurrent read
	// for i := 0; i < num_read; i++{
	// 	go read_load(i, rTime)
	// }

	// // retrieve elapsed time
	// write_times := make([]int, num_write)
	// for i := 0; i < num_write; i++{
	// 	millisec := int(<-wTime/time.Millisecond)
	// 	WTotal += millisec
	// 	write_times[i] = millisec
	// }
	// // retrieve elapsed time
	// read_times := make([]int, num_read)
	// for i := 0; i < num_read; i++{
	// 	millisec := int(<-rTime/time.Millisecond)
	// 	RTotal += millisec
	// 	read_times[i] = millisec
	// }

	sort.Ints(write_times)
	sort.Ints(read_times)

	fmt.Printf("Avg write time: %f ms\n", float64(WTotal)/float64(num_write))
	fmt.Printf("Avg read time: %f ms\n", float64(RTotal)/float64(num_read))
	fmt.Printf("95-th percentile for write time: %d ms\n", write_times[int(float64(num_write) * 0.95)])
	fmt.Printf("95-th percentile for read time: %d ms\n", read_times[int(float64(num_read) * 0.95)])
}
