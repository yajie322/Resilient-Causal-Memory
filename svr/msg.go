package main

import (
	"fmt"
	"bytes"
	"encoding/gob"	
)

type Message struct {
	Type 	int
	Id		int
	Key		int
	Val		string
	Vec		[]int
}

// encode msg
func getGobFromMsg(msg *Message) []byte {
	var res bytes.Buffer

	enc := gob.NewEncoder(&res)
	if err := enc.Encode(&msg); err != nil {
		fmt.Println(err)
	}
	return res.Bytes()
}

// decode msg
func getMsgFromGob(msgBytes []byte) Message {
	var buff bytes.Buffer
	var msg Message

	buff.Write(msgBytes)
	dec := gob.NewDecoder(&buff)
	if err := dec.Decode(&msg); err != nil {
		fmt.Println(err)
		msg.Type = ERROR
	}
	return msg
}

func (svr *Server)createRep(msg Message) *Message {
	var out Message
	switch msg.Type{
	case READ:
		res := svr.read(msg.Key)
		out = Message{Type: DATA, Id: node_id, Key: msg.Key, Val: res, Vec: make([]int, 1)}
	case WRITE:
		svr.write(msg.Key,msg.Val)
		out = Message{Type: ACK, Id: node_id, Key: msg.Key, Val: "", Vec: make([]int, 1)}
	}
	return &out
}