package main

import (
  "errors"
  "fmt"
  "strings"
)



// I wonder if using interface{} would also be ok?
type DumbQueueNode struct {
    value interface{}
    next *DumbQueueNode
}

// Dumb FIFO queue, because channels kind of suck for this.
type DumbQueue struct {
  head *DumbQueueNode
  tail *DumbQueueNode
  size int
}

func createDumbQueue() *DumbQueue {
  return &DumbQueue{nil, nil, 0}
}

func (dq *DumbQueue) isEmpty() bool {
  // Safety belt; guard against negative size...
  return dq.size <= 0
}

func (dq *DumbQueue) enqueue(val interface{}) {
  // Put a value at the end of the queue, using the tail pointer to jump to the end.
  newNode := &DumbQueueNode{val, nil}

  if dq.isEmpty() {
  // Special case for empty queue: update the head pointer as well as the tail pointer
    dq.head = newNode
  } else {
    // Get the existing tail
    dq.tail.next = newNode
  }
  // Set the tail to be the new node
  dq.tail = newNode
  // Increment the size
  dq.size++
}

// Remove the least-recently enqueued element from the queue and return it.
func (dq *DumbQueue) dequeue() (interface{}, error) {
  if dq.isEmpty() {
    return nil, errors.New("Cannot dequeue from empty queue")
  }
  returnNode := dq.head
  // Speical case: if there's only one element in the queue, then we have to set tail to nil too.
  if dq.size == 1 {
    // Assert invariants.
    if returnNode.next != nil {
      return nil, errors.New("Node of singleton queue has non-nil next pointer")
    }
    if dq.tail != returnNode {
      return nil, errors.New("Tail of singleton queue does not point to head")
    }
    dq.tail = nil
  }
  // Set head to the next node
  dq.head = returnNode.next
  dq.size--
  return returnNode.value, nil
}

// For printing to a string
type stringify func(interface{}) string

func (dq *DumbQueue) toString(printVal stringify) string {
  var sb strings.Builder
  sb.WriteString(fmt.Sprintf("DumbQueue{head:%p tail:%p size:%d", dq.head, dq.tail, dq.size))
  if dq.isEmpty() {
    sb.WriteString("}")
    return sb.String()
  }
  // else 
  sb.WriteString(" items:\n")
  for curNode := dq.head; curNode != nil; curNode = curNode.next {
    sb.WriteString("  ")
    sb.WriteString(printVal(curNode.value))
    sb.WriteString("\n")
  }
  sb.WriteString("}")
  return sb.String()
}