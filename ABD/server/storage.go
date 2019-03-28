package main

func set(tv TagVal) Message{
	if local,isIn := mData[tv.Key]; !isIn || local.Tag.smaller(tv.Tag) {
		mutex.Lock()
		mData[tv.Key] = tv
		mutex.Unlock()
	}

	return Message{OpType: SET, Tv: TagVal{Tag: Tag{Id: "", Ts: 0}, Key: 0, Val: ""}}
}

func get(tv TagVal) Message{
	var res Message
	res.OpType = GET
	if local,isIn := mData[tv.Key]; isIn{
		res.Tv = local
	} else {
		res.Tv = TagVal{Tag: Tag{Id: "", Ts: 0}, Key: tv.Key, Val: ""}
	}
	return res
}