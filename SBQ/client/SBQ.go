package main

import (
	"fmt"
)

// sbq write
func write(key int, val string){
	ts := getTs(key)
	tv := update(key,ts,val)
	store(tv)
}

// sbq read
func read(key int) string{
	dealer := createDealerSocket()
	defer dealer.Close()

	// init tagval
	tv := TagVal{Ts: -1, Key: key, Val: ""}
	msg := Message{OpType: GET, Tv: tv}
	broadcast(msg, dealer)

	responses := make(map[TagVal]int)
	// count received pairs
	for i := 0; i < READ_QUORUM; i++ {
		tmp := recvData(dealer)
		if _,isIn := responses[tmp]; isIn {
			responses[tmp] += 1
		} else {
			responses[tmp] = 1
		}
	}

	// select largest from A'
	for val,count := range responses{
		if count > F && val.Ts > tv.Ts{
			tv = val
		}
	}

	return tv.Val
}

// read ts
func getTs(key int) int{
	dealer := createDealerSocket()
	defer dealer.Close()

	msg := Message{OpType: GETTS, Tv: TagVal{Ts: 0, Key: key, Val: ""}}
	broadcast(msg, dealer)

	maxTs := -1
	for i := 0; i < READ_QUORUM; i++ {
		tmp := recvTs(dealer)
		if maxTs < tmp {
			maxTs = tmp
		}
	}

	return maxTs
}

// store
func store(tv TagVal){
	dealer := createDealerSocket()
	defer dealer.Close()
	fmt.Println("set ts", tv.Ts, "key",tv.Key)
	msg := Message{OpType: STORE, Tv: tv}
	sendStore(msg, dealer)
}

// get new ts
func update(key int, ts int, val string) TagVal{
	tv := TagVal{Ts:ts,Key:key,Val:val}
	if maxUsedTs < ts{
		tv.Ts += 1
	} else {
		tv.Ts = maxUsedTs + 1
	}
	maxUsedTs = tv.Ts
	return tv
}