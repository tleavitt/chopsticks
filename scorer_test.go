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

func TestScorePropagateScoresSimple(t *testing.T) {
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
  leaves := make(map[*playNode][]*playNode, 1)
  leaves[sonNode] = []*playNode{}
  if err := scorePlayGraph(leaves, make(map[*loopGraph]map[*playNode]int)); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(grandpaNode, t)
  fmt.Println("Grandpa: " + grandpaNode.toString())
  fmt.Println("Dad: " + dadNode.toString())
  fmt.Println("Son: " + sonNode.toString())

  fmt.Println("stopping TestPropagateScores")
}

func TestScorePropagateScoresFork(t *testing.T) {
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
  leaves := make(map[*playNode][]*playNode, 1)
  leaves[three] = []*playNode{}
  // Should require exactly two nodes on the frontier (two and two prime)
  if err := scorePlayGraph(leaves, make(map[*loopGraph]map[*playNode]int)); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(one, t)

  fmt.Println("stopping TestPropagateScoresFork")
}

func TestScorePropagateScoresLoop1(t *testing.T) {
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

  leaves := make(map[*playNode][]*playNode)

  loops := [][]*playNode{
    []*playNode{oneNode, twoNode, threeNode, fourNode},
  } 
  loopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(loopGraphs)

  if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode, t)

  fmt.Println("stopping TestPropagateScoresLoop")
}

func TestScorePropagateScoresLoop2(t *testing.T) {
  fmt.Println("starting TestPropagateScoresLoop2")

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

  exit := &gameState{
    player{0, 1}, player{0, 0}, Player2,
  }

  entryNode := createPlayNodeCopyGs(entry)
  oneNode := createPlayNodeCopyGs(one)
  twoNode := createPlayNodeCopyGs(two)
  threeNode := createPlayNodeCopyGs(three)
  exitNode := createPlayNodeCopyGs(exit)
  fourNode := createPlayNodeCopyGs(four)

  // Wire everything up, note that moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, move{Right, Left})
  addParentChildEdges(oneNode, twoNode, move{Right, Right})
  addParentChildEdges(twoNode, threeNode, move{Right, Right})
  addParentChildEdges(threeNode, fourNode, move{Right, Right})
  addParentChildEdges(threeNode, exitNode, move{Right, Left})
  addParentChildEdges(fourNode, oneNode, move{Right, Right})

  leaves := map[*playNode][]*playNode{
    exitNode: []*playNode{},
  }

  loops := [][]*playNode{
    []*playNode{oneNode, twoNode, threeNode, fourNode},
  } 
  loopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(loopGraphs)

  if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode, t)

  fmt.Println("stopping TestPropagateScoresLoop2")
}

func TestScorePropagateScoresComplex(t *testing.T) {
  fmt.Println("starting TestPropagateScoresComplex")

  entry := &gameState{
    player{1, 4}, player{0, 3}, Player2,
  }

  dad := &gameState{
    player{1, 2}, player{0, 3}, Player1,
  }

  bro := &gameState{
    player{1, 2}, player{0, 1}, Player2,
  }

  sis := &gameState{
    player{1, 2}, player{0, 2}, Player2,
  }

  sis2 := &gameState{
    player{1, 2}, player{0, 4}, Player2,
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

  exit := &gameState{
    player{0, 1}, player{0, 0}, Player2,
  }

  entryNode := createPlayNodeCopyGs(entry)
  oneNode := createPlayNodeCopyGs(one)
  twoNode := createPlayNodeCopyGs(two)
  threeNode := createPlayNodeCopyGs(three)
  exitNode := createPlayNodeCopyGs(exit)
  fourNode := createPlayNodeCopyGs(four)

  dadNode := createPlayNodeCopyGs(dad)
  broNode := createPlayNodeCopyGs(bro)
  sisNode := createPlayNodeCopyGs(sis)
  sis2Node := createPlayNodeCopyGs(sis2)

  // Wire everything up, note that moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, move{Right, Left})
  addParentChildEdges(oneNode, twoNode, move{Right, Right})
  addParentChildEdges(twoNode, threeNode, move{Right, Right})
  addParentChildEdges(threeNode, fourNode, move{Right, Right})
  addParentChildEdges(threeNode, exitNode, move{Right, Left})
  addParentChildEdges(fourNode, oneNode, move{Right, Right})

  addParentChildEdges(entryNode, dadNode, move{Left, Left})
  addParentChildEdges(dadNode, broNode, move{Left, Left})
  addParentChildEdges(dadNode, sisNode, move{Left, Right})
  addParentChildEdges(dadNode, sis2Node, move{Right, Left})

  leaves := map[*playNode][]*playNode{
    broNode: []*playNode{},
    sisNode: []*playNode{},
    sis2Node: []*playNode{},
    exitNode: []*playNode{},
  }

  loops := [][]*playNode{
    []*playNode{oneNode, twoNode, threeNode, fourNode},
  } 
  loopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(loopGraphs)

  if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode, t)

  fmt.Println("stopping TestPropagateScoresComplex")
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

func TestScoreSimpleLoop(t *testing.T) {
  fmt.Println("starting TestScoreSimpleLoop")
  // Note: no exit nodes here.
  loops := createSimpleLoop()
  distinctLoopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(distinctLoopGraphs)
  if err := scorePlayGraph(make(map[*playNode][]*playNode), loopGraphsToExitNodes); err != nil {
    t.Fatal(err.Error())
  }
  for _, loop := range loops {
    for _, node := range loop {
      if !node.isScored || node.score != 0 {
        t.Fatalf("Incorrectly scored node: %+v", node)
      }
    }
  }
  fmt.Println("stopping TestScoreSimpleLoop")
}
