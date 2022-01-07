package main

import (
  "testing"
  "fmt"
)

type Foo struct {
  bar string
  baz int
  other *Foo
}

func dequeueHandleErr(dq *DumbQueue, t *testing.T) *Foo {
  result, err := dq.dequeue()
  if err != nil {
    t.Fatal(err.Error())
  }
  return result.(*Foo)
}

func TestDumbQueue(t *testing.T) {
  foo1 := &Foo{"1", 1, nil,}
  foo2 := &Foo{"2", 2, foo1,}
  foo3 := &Foo{"3", 3, foo2,}

  dq := createDumbQueue()
  if dq.size != 0 || !dq.isEmpty() {
    t.Fatalf("Empty queue has size %d", dq.size)
  }

  dq.enqueue(foo1)
  if dq.head == nil || dq.head != dq.tail {
    t.Fatalf("Unexpected queue pointers: %+v", dq)
  }
  copyFoo1 := dequeueHandleErr(dq, t)
  if copyFoo1 != foo1 {
    t.Fatalf("Did not dequeue enqueued value: %p (should be %p)", copyFoo1, foo1)
  }
  if !dq.isEmpty() {
    t.Fatalf("Queue should be empty but has size %d", dq.size)
  }
  if dq.head != nil || dq.tail != nil {
    t.Fatalf("Unexpected queue pointers: %+v", dq)
  }

  // Interleave enqueues
  dq.enqueue(foo1)
  dq.enqueue(foo2)
  dequeueHandleErr(dq, t)
  copyFoo2 := dequeueHandleErr(dq, t)
  dq.enqueue(foo3)
  copyFoo3 := dequeueHandleErr(dq, t)

  if copyFoo3 != foo3 {
    t.Fatalf("Did not dequeue enqueued value: %p (should be %p)", copyFoo3, foo3)
  }

  copyFoo2.bar = "BAR"
  if foo2.bar != "BAR" {
    t.Fatalf("Change to dequeued value did not impact original value: original %+v, pointer %+v", foo2, copyFoo2)
  }

  _, err := dq.dequeue()
  if err == nil {
    t.Fatalf("Expected err from dequeueing empty queue")
  }
}

func fooToString(fi interface{}) string {
  foo := fi.(*Foo)
  return fmt.Sprintf("%+v", foo)
}

func TestDumbQueueString(t *testing.T) {
  foo1 := &Foo{"1", 1, nil,}
  foo2 := &Foo{"2", 2, foo1,}
  foo3 := &Foo{"3", 3, foo2,}

  dq := createDumbQueue()
  fmt.Println("Empty:")
  fmt.Printf("%s\n", dq.toString(fooToString))
  dq.enqueue(foo1)
  dq.enqueue(foo2)
  dq.enqueue(foo3)

  fmt.Println("Three values:")
  fmt.Printf("%s\n", dq.toString(fooToString))

  foo2.bar = "UPDATED"
  dq.dequeue()

  fmt.Println("After update:")
  fmt.Printf("%s\n", dq.toString(fooToString))
}
