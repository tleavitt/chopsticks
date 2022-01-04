package main

import (
  "fmt"
  "testing"
)

func TestLoopsSimple(t *testing.T) {
  fmt.Println("starting TestSimpleLoops")
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}
  loops := [][]*playNode{
    []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1)},
    []*playNode{createPlayNodeCopyGs(gs2), createPlayNodeCopyGs(gs2)},
    []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs2)},
  } 
  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  for lg, _ := range distinctLoopGraphs {
    var curNode *loopNode
    var i int
    for i, curNode = 0, lg.head; i == 0 || curNode != lg.head; i++ {
      if curNode.lg != lg {
        t.Fatalf("Loop node contained in loop graph does not point to it: %+v", curNode)
      }
      if curNode.pn == nil || curNode.pn.ln == nil {
        t.Fatalf("Loop node pointers not set correctly: %+v", curNode)
      }
      if curNode.pn.ln != curNode {
        t.Fatalf("Play node pointer is incorrect: %+v", curNode.pn)
      }
      // Update curNode
      if len(curNode.nextNodes) != 1 {
        t.Fatalf("Unexpected next nodes: %+v", curNode)
      }
      for nextNode, _ := range curNode.nextNodes {
        curNode = nextNode 
      }
    }
  }

  fmt.Println("finished TestSimpleLoops")
}


func assertParentChild(parent *loopNode, child *loopNode, t *testing.T) {
  fmt.Println("assertParentChild")
  if parent == nil {
    t.Fatal("Parent is nil")
  }
  if child == nil {
    t.Fatal("Child is nil")
  }
  if !parent.nextNodes[child] {
    t.Fatalf("parent %+v does not contain child %+v", parent, child)
  }
  if !child.prevNodes[parent] {
    t.Fatalf("child %+v does not contain parent %+v", child, parent)
  }
}

func TestLoopsInterlinked(t *testing.T) {
  fmt.Println("starting TestLoopsInterlinked")
  gs := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  commonNode1 := createPlayNodeCopyGs(gs)
  commonNode2 := createPlayNodeCopyGs(gs)

  pn11 := createPlayNodeCopyGs(gs)
  pn12 := createPlayNodeCopyGs(gs)
  pn14 := createPlayNodeCopyGs(gs)

  pn22 := createPlayNodeCopyGs(gs)
  pn23 := createPlayNodeCopyGs(gs)

  pn31 := createPlayNodeCopyGs(gs)
  pn33 := createPlayNodeCopyGs(gs)
  pn34 := createPlayNodeCopyGs(gs)

  loops := [][]*playNode{
    []*playNode{pn11, pn12, commonNode1, pn14},
    []*playNode{commonNode1, pn22, pn23, commonNode2},
    []*playNode{pn31, commonNode2, pn33, pn34},
  } 

  fmt.Println(len(loops))

  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  if len(distinctLoopGraphs) != 1 {
    t.Fatal("Did not join distinct loops into one")
  }

  for _, loop := range loops {
    for _, curNode := range loop {
      if curNode.ln == nil {
        t.Fatalf("Loop node pointers not set correctly: %+v", curNode)
      } else {
        fmt.Printf("%+v\n", curNode.ln)
      }
    }
  }

  // Common node 1
  assertParentChild(pn12.ln, commonNode1.ln, t)
  assertParentChild(commonNode1.ln, pn14.ln, t)
  assertParentChild(commonNode2.ln, commonNode1.ln, t)
  assertParentChild(commonNode1.ln, pn22.ln, t)

  // Common node 2
  assertParentChild(pn23.ln, commonNode2.ln, t)
  assertParentChild(pn31.ln, commonNode2.ln, t)
  assertParentChild(commonNode2.ln, pn33.ln, t)

  fmt.Println("finished TestLoopsInterlinked")
}