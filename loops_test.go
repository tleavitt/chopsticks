package main

import (
  "fmt"
  "testing"
)

func TestSimpleLoops(t *testing.T) {
  fmt.Println("starting TestSimpleLoops")
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}
  loops := map[gameState][]*playNode{
    *initGame(): []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1)},
    *initGame(): []*playNode{createPlayNodeCopyGs(gs2), createPlayNodeCopyGs(gs2)},
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
    }
  }

  fmt.Println("finished TestSimpleLoops")
}