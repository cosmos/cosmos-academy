package types

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"container/heap"
)

func TestAddPopUpdate(t *testing.T) {
	items := map[string]int{
		"banana": 6, "apple": 2, "pear": 5, "tomato": 3,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			Value:    value,
			Priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	// Insert a new item and then modify its priority.
	item := &Item{
		Value:    "orange",
		Priority: 1,
	}
	heap.Push(&pq, item)

	assert.Equal(t, "orange", pq.Peek().Value, "Peek does not work")

	pq.Update(item.Value, 5)

	pq.Remove("tomato")

	expected := []string{"apple", "orange", "pear", "banana"}

	var actual []string

	// Take the items out; they arrive in increasing priority order.
	for pq.Len() > 0 {
		actual = append(actual, heap.Pop(&pq).(*Item).Value)
	}
	
	assert.Equal(t, expected, actual, "Push and pop do not work")
}