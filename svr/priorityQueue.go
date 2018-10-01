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
	vec_x := pq[i].Vec
	vec_y := pq[j].Vec
	flag := true
	tmp := 0
	for k:= 0; k < len(vec_x); k++ {
		if vec_x[k] > vec_y[k]{
			flag = false
			break
		} else if vec_x[k] == vec_y[k]{
			tmp += 1
		}
	}
	if tmp == len(vec_x) {
		flag = false
	}
	return flag
}

func (pq PriorityQueue) Swap(i,j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Message)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}