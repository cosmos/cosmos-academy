package types

import (
	"container/heap"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Value should be identifier of listing
type Item struct {
	Value    string // The Value of the item; arbitrary.
	Priority int    // The Priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Enforce deterministic ordering with unique values
	if pq[i].Priority == pq[j].Priority {
		return pq[i].Value < pq[j].Value	
	}
	return pq[i].Priority < pq[j].Priority
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

// update modifies the Priority and Value of an Item in the queue.
func (pq *PriorityQueue) Update(Value string, Priority int) sdk.Error {
	for _, item := range *pq {
		if item.Value == Value {
			item.Priority = Priority
			heap.Fix(pq, item.index)
			return nil
		}
	}
	return ErrInvalidBallot(2, "Given identifier not found in queue")
}

func (pq *PriorityQueue) Remove(Value string) sdk.Error {
	for _, item := range *pq {
		if item.Value == Value {
			heap.Remove(pq, item.index)
			return nil
		}
	}
	return ErrInvalidBallot(2, "Given identifier not found in queue")
}