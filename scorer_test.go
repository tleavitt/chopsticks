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
  fmt.Println("starting TestMostWinningNodesInterlinked")
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}

  // Odd numbers are player1's turn, even numbers are player 2's turn
  commonNode1 := createPlayNodeCopyGs(gs1) // 1
  commonNode2 := createPlayNodeCopyGs(gs2) // 2

  pn11 := createPlayNodeCopyGs(gs1) // 1
  pn12 := createPlayNodeCopyGs(gs2) // 2
  pn14 := createPlayNodeCopyGs(gs2) // 2

  pn22 := createPlayNodeCopyGs(gs2) // 2
  pn23 := createPlayNodeCopyGs(gs1) // 1

  pn31 := createPlayNodeCopyGs(gs1) // 1
  pn33 := createPlayNodeCopyGs(gs1) // 1
  pn34 := createPlayNodeCopyGs(gs2) // 2

  loops := [][]*playNode{
    []*playNode{pn11, pn12, commonNode1, pn14},
    []*playNode{commonNode1, pn22, pn23, commonNode2},
    []*playNode{pn31, commonNode2, pn33, pn34},
  } 

  // ExitNode 1: player1 wins
  exitNode1 := createPlayNodeReuseGs(&gameState{
    player{0, 2,}, player{0, 0,}, Player2,})
  exitNode1.score = 1
  exitNode1.isScored = true

  // ExitNode 2: player2 will win 
  exitNode2 := createPlayNodeReuseGs(&gameState{
    player{0, 2,}, player{2, 2,}, Player1,})
  exitNode2.score = -1
  exitNode2.isScored = true

  // Exit Node 3: player1 will win
  exitNode3 := createPlayNodeReuseGs(&gameState{
    player{0, 1,}, player{0, 1,}, Player1,})
  exitNode3.score = 1
  exitNode3.isScored = true

  // ExitNode 4: dead even
  exitNode4 := createPlayNodeReuseGs(&gameState{
    player{1, 1,}, player{1, 1,}, Player2,})
  exitNode4.score = 0
  exitNode4.isScored = true

  m := move{Left, Left} // The specific move doesn't matter, just the edges

  // One winning exit node for 23
  pn23.nextNodes[m] = exitNode1

  // One winning and one losing exit node for 34
  pn34.nextNodes[m] = exitNode2 //
  pn34.nextNodes[move{Left, Right}] = exitNode3 //

  // One neutral exit node for common1
  commonNode1.nextNodes[m] = exitNode4

  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  if len(distinctLoopGraphs) != 1 {
    t.Fatal("Did not join distinct loops into one")
  }

  for lg, _ := range distinctLoopGraphs {
    b1, b2, err := findMostWinningNodes(lg)
    if err != nil {
      t.Fatal(err.Error())
    }
    fmt.Printf("Most winning nodes: %+v, %+v", b1, b2)
    // Best player1 score should be p23
    if b1.score != 1 && b1.node.pn != pn23 {
      t.Fatalf("Unexpected winning node for player1: %+v, playNode %+v", b1, b1.node.pn)
    }

    // Best player2 score should be p34
    if b2.score != 1 && b2.node.pn != pn34 {
      t.Fatalf("Unexpected winning node for player2: %+v, playNode %+v", b1, b1.node.pn)
    }
  }

  fmt.Println("finished TestMostWinningNodesInterlinked")
}