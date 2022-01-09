package main

import (
  "fmt"
  "testing"
)


func ensureAllNodesScored(root *playNode, t *testing.T) {
  ensureAllNodesScoredImpl(root, t, make(map[gameState]bool))
}

func ensureAllNodesScoredImpl(root *playNode, t *testing.T, visitedStates map[gameState]bool) {
  if visitedStates[*root.gs] {
    return
  }
  visitedStates[*root.gs] = true
  if !root.isScored {
    t.Fatalf("Found unscored node: %s", root.toString())
  }
  for _, child := range root.nextNodes {
    ensureAllNodesScoredImpl(child, t, visitedStates)
  }
  return
}

func TestPropagateScores1(t *testing.T) {
  fmt.Println("starting TestPropagateScores")

  grandpa := &gameState{
    player{1, 1}, player{1, 1}, Player1,
  }
  dad := &gameState{
    player{1, 1}, player{1, 2}, Player2,
  }
  son := &gameState{
    player{0, 1}, player{1, 2}, Player1,
  }
  grandpaNode := createPlayNodeCopyGs(grandpa)
  dadNode := createPlayNodeCopyGs(dad)
  sonNode := createPlayNodeCopyGs(son)

  // Wire everything up
  addParentChildEdges(grandpaNode, dadNode, move{Left, Left})

  addParentChildEdges(dadNode, sonNode, move{Right, Left})

  // Score
  leaves := make(map[gameState]*playNode, 1)
  leaves[*sonNode.gs] = sonNode
  if err := scorePlayGraph(leaves, make(map[*loopGraph]bool)); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(grandpaNode, t)
  fmt.Println("Grandpa: " + grandpaNode.toString())
  fmt.Println("Dad: " + dadNode.toString())
  fmt.Println("Son: " + sonNode.toString())

  fmt.Println("stopping TestPropagateScores")
}

func TestPropagateScoresFork(t *testing.T) {
  fmt.Println("starting TestPropagateScoresFork")

  oneS := &gameState{
    player{1, 2}, player{1, 2}, Player1,
  }
  twoS := &gameState{
    player{1, 2}, player{1, 1}, Player2,
  }
  twoprimeS := &gameState{
    player{1, 2}, player{1, 5}, Player2,
  }
  threeS := &gameState{
    player{0, 1}, player{1, 2}, Player1,
  }

  one := createPlayNodeCopyGs(oneS)
  two := createPlayNodeCopyGs(twoS)
  twoprime := createPlayNodeCopyGs(twoprimeS)
  three := createPlayNodeCopyGs(threeS)

  // Wire everything up, note that moves don't actually matter here.
  addParentChildEdges(one, two, move{Right, Right})
  addParentChildEdges(one, twoprime, move{Right, Left})
  addParentChildEdges(two, three, move{Right, Right})
  addParentChildEdges(twoprime, three, move{Right, Left})

  // Score
  leaves := make(map[gameState]*playNode, 1)
  leaves[*three.gs] = three
  // Should require exactly two nodes on the frontier (two and two prime)
  if err := scorePlayGraph(leaves, make(map[*loopGraph]bool)); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(one, t)

  fmt.Println("starting TestPropagateScoresFork")
}

func TestPropagateScoresLoop(t *testing.T) {
  fmt.Println("starting TestPropagateScoresLoop")

  // The three-four loop:
  entry := &gameState{
    player{1, 4}, player{0, 3}, Player2,
  }
  // => RH->LH
  one := &gameState{
    player{0, 4}, player{0, 3}, Player1,
  } // All the rest are RH->RH
  two := &gameState{
    player{0, 4}, player{0, 2}, Player2,
  }
  three := &gameState{
    player{0, 1}, player{0, 2}, Player1,
  }
  four := &gameState{
    player{0, 1}, player{0, 3}, Player2,
  } // Then we loop back to one

  entryNode := createPlayNodeCopyGs(entry)
  oneNode := createPlayNodeCopyGs(one)
  twoNode := createPlayNodeCopyGs(two)
  threeNode := createPlayNodeCopyGs(three)
  fourNode := createPlayNodeCopyGs(four)

  // Wire everything up, note that moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, move{Right, Left})
  addParentChildEdges(oneNode, twoNode, move{Right, Right})
  addParentChildEdges(twoNode, threeNode, move{Right, Right})
  addParentChildEdges(threeNode, fourNode, move{Right, Right})
  addParentChildEdges(fourNode, oneNode, move{Right, Right})

  leaves := make(map[gameState]*playNode)

  loops := [][]*playNode{
    []*playNode{oneNode, twoNode, threeNode, fourNode},
  } 
  loopGraphs := createDistinctLoopGraphs(loops) 

  if err := scorePlayGraph(leaves, loopGraphs); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode, t)

  fmt.Println("starting TestPropagateScoresLoop")
}

func createSimpleLoop() [][]*playNode {
  gs1 := &gameState{player{1, 1,}, player{1, 2,}, Player1,}
  gs2 := &gameState{player{1, 1,}, player{2, 2,}, Player2,}
  // Note: no exit nodes here.
  loops := [][]*playNode{
    []*playNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs2)},
  } 
  return loops
}

func TestMostWinningNodesSimple(t *testing.T) {
  fmt.Println("starting TestMostWinningNodesSimple")
  // Note: no exit nodes here.
  loops := createSimpleLoop()
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

func TestScoreSimpleLoop(t *testing.T) {
  fmt.Println("starting TestScoreSimpleLoop")
  // Note: no exit nodes here.
  loops := createSimpleLoop()
  distinctLoopGraphs := createDistinctLoopGraphs(loops) 
  if err := scorePlayGraph(make(map[gameState]*playNode), distinctLoopGraphs); err != nil {
    t.Fatal(err.Error())
  }
  for _, loop := range loops {
    for _, node := range loop {
      if !node.isScored || node.score != 0 {
        t.Fatalf("Incorrectly scored node: %+v", node)
      }
    }
  }
}

func createInterlockedLoops() [][]*playNode {
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

  return loops
}

func TestMostWinningNodesInterlinked(t *testing.T) {
  fmt.Println("starting TestMostWinningNodesInterlinked")
  loops := createInterlockedLoops()
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
    pn23 := loops[1][2]
    if b1.score != 1 && b1.node.pn != pn23 {
      t.Fatalf("Unexpected winning node for player1: %+v, playNode %+v", b1, b1.node.pn)
    }

    // Best player2 score should be p34
    pn34 := loops[2][3]
    if b2.score != 1 && b2.node.pn != pn34 {
      t.Fatalf("Unexpected winning node for player2: %+v, playNode %+v", b1, b1.node.pn)
    }
  }

  fmt.Println("finished TestMostWinningNodesInterlinked")
}