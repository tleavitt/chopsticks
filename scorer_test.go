package main

import (
  "fmt"
  "testing"
)


func ensureAllNodesScored(root *PlayNode, t *testing.T) {
  ensureAllNodesScoredImpl(root, t, make(map[GameState]bool))
}

func ensureAllNodesScoredImpl(root *PlayNode, t *testing.T, visitedStates map[GameState]bool) {
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

  grandpa := &GameState{
    Player{1, 1}, Player{1, 1}, Player1,
  }
  dad := &GameState{
    Player{1, 1}, Player{1, 2}, Player2,
  }
  son := &GameState{
    Player{0, 1}, Player{1, 2}, Player1,
  }
  grandpaNode := createPlayNodeCopyGs(grandpa)
  dadNode := createPlayNodeCopyGs(dad)
  sonNode := createPlayNodeCopyGs(son)

  // Wire everything up
  addParentChildEdges(grandpaNode, dadNode, Move{Left, Left})

  addParentChildEdges(dadNode, sonNode, Move{Right, Left})

  // Score
  leaves := make(map[*PlayNode][]*PlayNode, 1)
  leaves[sonNode] = []*PlayNode{}
  if err := scorePlayGraph(leaves, make(map[*loopGraph]map[*PlayNode]int)); err != nil {
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

  oneS := &GameState{
    Player{1, 2}, Player{1, 2}, Player1,
  }
  twoS := &GameState{
    Player{1, 2}, Player{1, 1}, Player2,
  }
  twoprimeS := &GameState{
    Player{1, 2}, Player{1, 5}, Player2,
  }
  threeS := &GameState{
    Player{0, 1}, Player{1, 2}, Player1,
  }

  one := createPlayNodeCopyGs(oneS)
  two := createPlayNodeCopyGs(twoS)
  twoprime := createPlayNodeCopyGs(twoprimeS)
  three := createPlayNodeCopyGs(threeS)

  // Wire everything up, note that Moves don't actually matter here.
  addParentChildEdges(one, two, Move{Right, Right})
  addParentChildEdges(one, twoprime, Move{Right, Left})
  addParentChildEdges(two, three, Move{Right, Right})
  addParentChildEdges(twoprime, three, Move{Right, Left})

  // Score
  leaves := make(map[*PlayNode][]*PlayNode, 1)
  leaves[three] = []*PlayNode{}
  // Should require exactly two nodes on the frontier (two and two prime)
  if err := scorePlayGraph(leaves, make(map[*loopGraph]map[*PlayNode]int)); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(one, t)

  fmt.Println("stopping TestPropagateScoresFork")
}

func TestScorePropagateScoresLoop1(t *testing.T) {
  fmt.Println("starting TestPropagateScoresLoop")

  // The three-four loop:
  entry := &GameState{
    Player{1, 4}, Player{0, 3}, Player2,
  }
  // => RH->LH
  one := &GameState{
    Player{0, 4}, Player{0, 3}, Player1,
  } // All the rest are RH->RH
  two := &GameState{
    Player{0, 4}, Player{0, 2}, Player2,
  }
  three := &GameState{
    Player{0, 1}, Player{0, 2}, Player1,
  }
  four := &GameState{
    Player{0, 1}, Player{0, 3}, Player2,
  } // Then we loop back to one

  entryNode := createPlayNodeCopyGs(entry)
  oneNode := createPlayNodeCopyGs(one)
  twoNode := createPlayNodeCopyGs(two)
  threeNode := createPlayNodeCopyGs(three)
  fourNode := createPlayNodeCopyGs(four)

  // Wire everything up, note that Moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, Move{Right, Left})
  addParentChildEdges(oneNode, twoNode, Move{Right, Right})
  addParentChildEdges(twoNode, threeNode, Move{Right, Right})
  addParentChildEdges(threeNode, fourNode, Move{Right, Right})
  addParentChildEdges(fourNode, oneNode, Move{Right, Right})

  leaves := make(map[*PlayNode][]*PlayNode)

  loops := [][]*PlayNode{
    []*PlayNode{oneNode, twoNode, threeNode, fourNode},
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
  entry := &GameState{
    Player{1, 4}, Player{0, 3}, Player2,
  }
  // => RH->LH
  one := &GameState{
    Player{0, 4}, Player{0, 3}, Player1,
  } // All the rest are RH->RH
  two := &GameState{
    Player{0, 4}, Player{0, 2}, Player2,
  }
  three := &GameState{
    Player{0, 1}, Player{0, 2}, Player1,
  }
  four := &GameState{
    Player{0, 1}, Player{0, 3}, Player2,
  } // Then we loop back to one

  exit := &GameState{
    Player{0, 1}, Player{0, 0}, Player2,
  }

  entryNode := createPlayNodeCopyGs(entry)
  oneNode := createPlayNodeCopyGs(one)
  twoNode := createPlayNodeCopyGs(two)
  threeNode := createPlayNodeCopyGs(three)
  exitNode := createPlayNodeCopyGs(exit)
  fourNode := createPlayNodeCopyGs(four)

  // Wire everything up, note that Moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, Move{Right, Left})
  addParentChildEdges(oneNode, twoNode, Move{Right, Right})
  addParentChildEdges(twoNode, threeNode, Move{Right, Right})
  addParentChildEdges(threeNode, fourNode, Move{Right, Right})
  addParentChildEdges(threeNode, exitNode, Move{Right, Left})
  addParentChildEdges(fourNode, oneNode, Move{Right, Right})

  leaves := map[*PlayNode][]*PlayNode{
    exitNode: []*PlayNode{},
  }

  loops := [][]*PlayNode{
    []*PlayNode{oneNode, twoNode, threeNode, fourNode},
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

  entry := &GameState{
    Player{1, 4}, Player{0, 3}, Player2,
  }

  dad := &GameState{
    Player{1, 2}, Player{0, 3}, Player1,
  }

  bro := &GameState{
    Player{1, 2}, Player{0, 1}, Player2,
  }

  sis := &GameState{
    Player{1, 2}, Player{0, 2}, Player2,
  }

  sis2 := &GameState{
    Player{1, 2}, Player{0, 4}, Player2,
  }

  // => RH->LH
  one := &GameState{
    Player{0, 4}, Player{0, 3}, Player1,
  } // All the rest are RH->RH
  two := &GameState{
    Player{0, 4}, Player{0, 2}, Player2,
  }
  three := &GameState{
    Player{0, 1}, Player{0, 2}, Player1,
  }
  four := &GameState{
    Player{0, 1}, Player{0, 3}, Player2,
  } // Then we loop back to one

  exit := &GameState{
    Player{0, 1}, Player{0, 0}, Player2,
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

  // Wire everything up, note that Moves don't actually matter here.
  addParentChildEdges(entryNode, oneNode, Move{Right, Left})
  addParentChildEdges(oneNode, twoNode, Move{Right, Right})
  addParentChildEdges(twoNode, threeNode, Move{Right, Right})
  addParentChildEdges(threeNode, fourNode, Move{Right, Right})
  addParentChildEdges(threeNode, exitNode, Move{Right, Left})
  addParentChildEdges(fourNode, oneNode, Move{Right, Right})

  addParentChildEdges(entryNode, dadNode, Move{Left, Left})
  addParentChildEdges(dadNode, broNode, Move{Left, Left})
  addParentChildEdges(dadNode, sisNode, Move{Left, Right})
  addParentChildEdges(dadNode, sis2Node, Move{Right, Left})

  leaves := map[*PlayNode][]*PlayNode{
    broNode: []*PlayNode{},
    sisNode: []*PlayNode{},
    sis2Node: []*PlayNode{},
    exitNode: []*PlayNode{},
  }

  loops := [][]*PlayNode{
    []*PlayNode{oneNode, twoNode, threeNode, fourNode},
  } 
  loopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(loopGraphs)

  if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode, t)

  fmt.Println("stopping TestPropagateScoresComplex")
}


func createSimpleLoop() [][]*PlayNode {
  gs1 := &GameState{Player{1, 1,}, Player{1, 2,}, Player1,}
  gs2 := &GameState{Player{1, 1,}, Player{2, 2,}, Player2,}
  // Note: no exit nodes here.
  loops := [][]*PlayNode{
    []*PlayNode{createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs1), createPlayNodeCopyGs(gs2)},
  } 
  return loops
}

func TestScoreSimpleLoop(t *testing.T) {
  fmt.Println("starting TestScoreSimpleLoop")
  // Note: no exit nodes here.
  loops := createSimpleLoop()
  distinctLoopGraphs := createLoopGraphs(loops) 
  loopGraphsToExitNodes := getAllExitNodes(distinctLoopGraphs)
  if err := scorePlayGraph(make(map[*PlayNode][]*PlayNode), loopGraphsToExitNodes); err != nil {
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
