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
func (clt *Client) write_load(num int, val string) time.Duration {
	// write data in the form (int, string) into table tmp
	// mutex.Lock()
	start := time.Now()
	clt.write(num, val)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	return elapsed
}

// read info from table by key
func (clt *Client) read_load(num int) time.Duration {
	// write data in the form table tmp with key = num	
	// mutex.Lock()
	start := time.Now()
	clt.read(num)
	end := time.Now()
	// mutex.Unlock()
	elapsed := end.Sub(start)
	return elapsed
}

func (clt *Client) workload(num int){
	var write_times []int
	var read_times []int

	num_read := 0
	num_write := 0

	var WTotal, RTotal int = 0, 0

	// insert value into table before start testing
	clt.initWrite(10)

	for i := 0; i < num; i++ {
		temp := rand.Float64()
		if temp < READ_PORTION {
			millisec := int(clt.read_load(i%10).Nanoseconds()/1000)
			RTotal += millisec
			read_times = append(read_times, millisec)
			num_read += 1
		} else {
			millisec := int(clt.write_load(i%10, strings.Repeat(strconv.Itoa(i % 10), DATA_SIZE)).Nanoseconds()/1000)
			WTotal += millisec
			write_times = append(write_times, millisec)
			num_write += 1
		}
	}

	sort.Ints(write_times)
	sort.Ints(read_times)

	fmt.Printf("Avg write time: %f ms\n", float64(WTotal)/float64(num_write))
	fmt.Printf("Avg read time: %f ms\n", float64(RTotal)/float64(num_read))
	fmt.Printf("95-th percentile for write time: %d ms\n", write_times[int(float64(num_write) * 0.95)])
	fmt.Printf("95-th percentile for read time: %d ms\n", read_times[int(float64(num_read) * 0.95)])
}
