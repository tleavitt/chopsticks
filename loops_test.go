package main

import (
  "fmt"
  "testing"
)

func TestLoopsSimple(t *testing.T) {
  fmt.Println("starting TestSimpleLoops")
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}
  loops := map[gameState][]*playNode{
    *initGame(): []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1)},
    *initGame(): []*playNode{createPlayNodeCopyGs(gs2), createPlayNodeCopyGs(gs2)},
    *initGame(): []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs2)},
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
  if !parent.nextNodes[child] {
    t.Fatalf("parent %+v does not contain child %+v", parent, child)
  }
  if !child.prevNodes[parent] {
    t.Fatalf("child %+v does not contain parent %+v", child, parent)
  }
}

func TestLoopsInterlinked(t *testing.T) {
  fmt.Println("starting TestSimpleLoops")
  gs := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  commonNode1 := createPlayNodeCopyGs(gs)
  commonNode2 := createPlayNodeCopyGs(gs)

  ln11 := createPlayNodeCopyGs(gs)
  ln12 := createPlayNodeCopyGs(gs)
  ln14 := createPlayNodeCopyGs(gs)

  ln22 := createPlayNodeCopyGs(gs)
  ln23 := createPlayNodeCopyGs(gs)

  ln31 := createPlayNodeCopyGs(gs)
  ln33 := createPlayNodeCopyGs(gs)
  ln34 := createPlayNodeCopyGs(gs)

  loops := map[gameState][]*playNode{
    *initGame(): []*playNode{ln11, ln12, commonNode1, ln14},
    *initGame(): []*playNode{commonNode1, ln22, ln23, commonNode2},
    *initGame(): []*playNode{ln31, commonNode2, ln33, ln34},
  } 

  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  if len(distinctLoopGraphs) != 1 {
    t.Fatal("Did not join distinct loops into one")
  }

  // Common node 1
  assertParentChild(ln12, commonNode1, t)
  assertParentChild(commonNode1, ln14, t)
  assertParentChild(commonNode2, commonNode1, t)
  assertParentChild(commonNode1, ln22, t)

  // Common node 2
  assertParentChild(ln23, commonNode2, t)
  assertParentChild(ln31, commonNode2, t)
  assertParentChild(commonNode2, ln33, t)

  fmt.Println("finished TestSimpleLoops")
}