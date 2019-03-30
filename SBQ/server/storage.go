package main

func store(tv TagVal) {
	mutex.Lock()
	mData[tv.Key] = tv
	mutex.Unlock()
}

func get(tv TagVal) Message{
	var res Message
	res.OpType = GET
	if local,isIn := mData[tv.Key]; isIn{
		res.Tv = local
	} else {
		res.Tv = TagVal{Ts: -1, Key: tv.Key, Val: ""}
	}
	return res
}

func getTs(tv TagVal) Message{
	var res Message
	res.OpType = GETTS
	res.Tv = TagVal{Ts: -1, Key: tv.Key, Val: ""}
	if local,isIn := mData[tv.Key]; isIn{
		res.Tv.Ts = local.Ts
	}
	return res
}