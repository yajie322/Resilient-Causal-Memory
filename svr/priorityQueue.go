// priority queue implementation
package main 

// import(
// 	"container/heap"
// )


type PriorityQueue []*Message

func (pq PriorityQueue) Len() int{
	return len(pq)
}

func (pq PriorityQueue) Less(i,j int) bool {
	vecX := pq[i].Vec
	vecY := pq[j].Vec
	tmp := 0
	for k:= 0; k < len(vecX); k++ {
		if vecX[k] > vecY[k]{
			return false
		} else if vecX[k] == vecY[k]{
			tmp += 1
		}
	}
	return tmp != len(vecX)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Message)
	*pq = append(*pq, item)
}

func (pq PriorityQueue) Swap(i,j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq PriorityQueue) Peek() *Message {
	return pq[0]
}