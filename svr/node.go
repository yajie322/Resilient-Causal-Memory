package main 

type Node struct {
	m_data		map[int]string
	vec_clock	[4]int
	outQueue	PriorityQueue
	inQueue		PriorityQueue
}

// initialize the node
func (n *Node) init(id int, group_size int){
	// init data as key(int)-value(string) pair
	n.m_data = make(map[int] string)
	// init vector timestamp with length group_size
	n.vec_clock = [group_size]
	// set vector timestamp to zero
	for i:= 0; i < group_size; i++{
		n.vec_clock = 0
	}
	// init priority queue outQueue
	n.outQueue = make(PriorityQueue,0)
	heap.Init(&n.outQueue)
	// init priority queue inQueue
	n.inQueue = make(PriorityQueue,0)
	heap.Init(&n.inQueue)
}

// perform read(key int), return value string
func (n *Node) read(key int) string{
	return n.m_data[key]
}

// perform write(id int, key int, value string)
func (n *Node) write(key int, value string) {
	// update vector clock
	n.vec_clock[id] += 1
	// update key-value pair
	n.m_data[key] = value
	// create Message object and push to outQueue
	msg := Message{Id: n.id, Key: key, Val:value,Vec: n.vec_clock}
	item := &Item{value: msg}
	heap.Push(&n.outQueue, item)
}

// To Do
func (n *Node) apply(){
	for n.inQueue.Len() > 0 {
		// To Do
		// item := heap.Pop(&n.inQueue).(*Item)
		// vec := item.value.Vec
	}
}