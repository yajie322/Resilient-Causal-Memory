package main 

// abd write
func write(key int, val string){
	tv := get(key)
	tv.update(val)
	set(tv)
}

// abd read
func read(key int) string{
	tv := get(key)
	set(tv)
	return tv.Val
}

// get phase
func get(key int) TagVal {
	dealer := createDealerSocket()
	defer dealer.Close()

	// init tagval
	tv := TagVal{Tag: Tag{Id: "", Ts: -1}, Key: key, Val: ""}
	msg := Message{OpType: GET, Tv: tv}
	sendToServer(msg, dealer)
	
	// find tagval with biggest tag in quorum 
	for i := 0; i < len(servers)/2 + 1; i++ {
		tmp := recvData(dealer)
		if tv.Tag.smaller(tmp.Tag) {
			tv = tmp
		}
	}
	return tv	
}

// set phase
func set(tv TagVal){
	dealer := createDealerSocket()
	defer dealer.Close()

	msg := Message{OpType: SET, Tv: tv}
	sendToServer(msg, dealer)
	// recv ack from quorum
	for i := 0; i < len(servers)/2 + 1; i++ {
		recvAck(dealer)
	}
}
