package types

import (
	"container/heap"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// value should be identifier of listing
type Item struct {
	value    string // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Enforce deterministic ordering with unique values
	if pq[i].priority == pq[j].priority {
		return pq[i].value < pq[j].value	
	}
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() Item {
	return *((*pq)[0])
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) Update(value string, priority int) sdk.Error {
	for _, item := range *pq {
		if item.value == value {
			item.priority = priority
			heap.Fix(pq, item.index)
			return nil
		}
	}
	return ErrInvalidBallot(2, "Given identifier not found in queue")
}

func (pq *PriorityQueue) Remove(value string) sdk.Error {
	for _, item := range *pq {
		if item.value == value {
			heap.Remove(pq, item.index)
			return nil
		}
	}
	return ErrInvalidBallot(2, "Given identifier not found in queue")
}