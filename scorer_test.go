package main

import (
  "fmt"
  "testing"
)

func TestMostWinningNodesSimple(t *testing.T) {
  fmt.Println("starting TestMostWinningNodesSimple")
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}
  // Note: no exit nodes here.
  loops := [][]*playNode{
    []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs2)},
  } 
  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  if len(distinctLoopGraphs) != 1 {
    t.Fatalf("Unexpected number of loop graphs: %d", len(distinctLoopGraphs))
  }

  for lg, _ := range distinctLoopGraphs {
    b1, b2, err := findMostWinningNodes(lg)
    if err != nil {
      t.Fatal(err.Error())
    }
    if b1.score != -2 || b2.score != -2 {
      t.Fatalf("Did not leave best moves uninitialized for simpler loop, b1: %+v, b2: %+v", b1, b2)
    }
  }


  fmt.Println("finished TestMostWinningNodesSimple")
}

func TestMostWinningNodesInterlinked(t *testing.T) {
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