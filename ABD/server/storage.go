package main

func set(tv TagVal) Message{
	flag := DEC
	tagVal := TagVal{Tag: Tag{Id: "", Ts: 0}, Key: 0, Val: ""}

	if local,isIn := mData[tv.Key]; !isIn || local.Tag.smaller(tv.Tag) {
		flag = SET
		mutex.Lock()
		mData[tv.Key] = tv
		mutex.Unlock()
	}

	return Message{OpType: flag, Tv:tagVal}
}

func get(tv TagVal) Message{
	var res Message
	res.OpType = GET
	if tv,isIn := mData[tv.Key]; isIn{
		res.Tv = tv
	} else {
		res.Tv = TagVal{Tag: Tag{Id: "", Ts: 0}, Key: tv.Key, Val: ""}
	}
	return res
}