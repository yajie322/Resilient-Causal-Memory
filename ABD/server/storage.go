package main

func set(tv TagVal) Message{
	var res Message

	state := mData[tv.Key].Tag
	if state.smaller(tv.Tag){
		mutex.Lock()
		mData[tv.Key] = tv
		mutex.Unlock()
		res.OpType = SET
		res.Tv = TagVal{Tag: state, Key: 0, Val: ""}
	} else {
		res = Message{OpType: DEC, 	Tv:	TagVal{Tag: state, Key: 0, Val: ""}}
	}
	return res
}

func get(tv TagVal) Message{
	var res Message
	res.OpType = GET
	res.Tv = mData[tv.Key]
	return res
}