package main

import (
  "fmt"
  "testing"
)

func testExploreStates(numFingers int8, maxDepth int, t *testing.T) {
  fmt.Println("starting TestExploreStates")
  prevNumFingers := setNumFingers(numFingers)
  startState := &gameState{
    player{1, 1}, player{1, 1}, Player1,
  }
  visitedStates := make(map[gameState]*playNode, 38)
  startNode, leaves, loops, err := exploreStates(createPlayNodeCopyGs(startState), visitedStates, maxDepth)
  if err != nil {
    t.Fatal(err)
  }
  fmt.Printf("Num states; %d\n", len(visitedStates))
  fmt.Printf("Leaves: %+v\n", leaves)
  fmt.Printf("Loops: %+v\n", loops)
  fmt.Println("Loops:")
  for _, loop := range loops {
    var prevTurn Turn
    for i, node := range loop {
      if i > 0 {
        if node.gs.turn == prevTurn {
          t.Fatalf("Consecutive loop members have the same turn") 
        }
      }
      fmt.Printf("%s, ", node.gs.toString())
      prevTurn = node.gs.turn
    }
    fmt.Printf("\n")
  }
  // Not necessarily the case with loops
  for leafNode, _ := range leaves {
    if len(leafNode.nextNodes) > 0 {
      t.Fatalf("Leaf node has children: %s", leafNode.toString())
    }
  }

  // Yay it's fixed now
  minDepth, maxDepth, err := startNode.validateEdges(true)
  if err != nil {
    t.Fatal(err)
  }

  fmt.Printf("Min game tree depth: %d, max game tree depth: %d\n", minDepth, maxDepth)

  fmt.Println(startNode.toTreeString(minDepth + 1))
  setNumFingers(prevNumFingers)
  fmt.Println("finished TestExploreStates")
}

func TestExploreStates3(t *testing.T) {
  testExploreStates(3, 17, t)
}

func TestExploreStates4(t *testing.T) {
  testExploreStates(4, 15, t)
}

func TestExploreStates5(t *testing.T) {
  testExploreStates(5, 20, t)
}


func TestExploreLoop(t *testing.T) {
  fmt.Println("starting TestExploreLoop")
  prevNumFingers := setNumFingers(5)

  startState := &gameState{
    player{0, 4}, player{0, 3}, Player1,
  }
  visitedStates := make(map[gameState]*playNode, 38)
  startNode, leaves, loops, err := exploreStates(createPlayNodeCopyGs(startState), visitedStates, 15)
  if err != nil {
    t.Fatal(err)
  }
  fmt.Printf("Num states; %d\n", len(visitedStates))
  fmt.Printf("Leaves: %+v\n", leaves)
  fmt.Printf("Loops: %+v\n", loops)
  fmt.Println("Loops:")
  for _, loop := range loops {
    var prevTurn Turn
    for i, node := range loop {
      if i > 0 {
        if node.gs.turn == prevTurn {
          t.Fatalf("Consecutive loop members have the same turn") 
        }
      }
      fmt.Printf("%s, ", node.gs.toString())
      prevTurn = node.gs.turn
    }
    fmt.Printf("\n")
  }
  // Not necessarily the case with loops
  // for _, leafNode := range leaves {
  //   if !leafNode.isTerminal() || len(leafNode.nextNodes) > 0 {
  //     t.Fatalf("Non-terminal leaf node: %s", leafNode.toString())
  //   }
  // }

  // Yay it's fixed now
  minDepth, maxDepth, err := startNode.validateEdges(true)
  if err != nil {
    t.Fatal(err)
  }

  fmt.Printf("Min game tree depth: %d, max game tree depth: %d\n", minDepth, maxDepth)

  // fmt.Println(startNode.toTreeString(15))
  setNumFingers(prevNumFingers)
  fmt.Println("finished TestExploreLoop")
}

func expectInvalidGraph(startNode *playNode, t *testing.T) {
  var err error
  if _, _, err = startNode.validateEdges(true); err == nil {
    t.Fatal("Expected validateEdges to error on invalid graph, but it did not")
  }
  fmt.Printf("Validate edges caught invalid graph: %s\n", err.Error())
}

func TestExploreInvalidGraphParent(t *testing.T) {
  fmt.Println("starting TestInvalidGraphParent")
  startState := &gameState{
    player{1, 1}, player{1, 1}, Player1,
  }
  nextState := &gameState{
    player{1, 1}, player{1, 2}, Player2,
  }
  startNode := createPlayNodeCopyGs(startState)
  nextNode := createPlayNodeCopyGs(nextState)
  // Only put one edge between startNode and nextNode
  startNode.nextNodes[move{Left, Left}] = nextNode
  expectInvalidGraph(startNode, t)
  fmt.Println("finished TestInvalidGraphParent")
}

func TestExploreInvalidGraphChild(t *testing.T) {
  fmt.Println("starting TestInvalidGraphChild")
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
  // Grandpa node does not point to dad node:
  addChildEdge(grandpaNode, dadNode)

  addParentChildEdges(dadNode, sonNode, move{Right, Left})

  // Should detect missing edge starting from son
  expectInvalidGraph(sonNode, t)
  fmt.Println("finished TestInvalidGraphChild")
}

func TestSolidifyScore(t *testing.T) {
  fmt.Println("starting TestSolidifyScore")
  gs1 := &gameState{
    player{1, 1}, player{1, 1}, Player1,
  }
  gs2 := &gameState{
    player{1, 1}, player{1, 2}, Player2,
  }
  gs3 := &gameState{
    player{0, 1}, player{1, 2}, Player2,
  }
  gs4 := &gameState{
    player{0, 1}, player{1, 2}, Player1,
  }
  gs5 := &gameState{
    player{0, 1}, player{0, 1}, Player2,
  }
  gs6 := &gameState{
    player{0, 1}, player{0, 0}, Player2,
  }

  n1 := createPlayNodeCopyGs(gs1)
  n2 := createPlayNodeCopyGs(gs2)
  n3 := createPlayNodeCopyGs(gs3)
  n4 := createPlayNodeCopyGs(gs4)
  n5 := createPlayNodeCopyGs(gs5)
  n6 := createPlayNodeCopyGs(gs6)

  // Add edges
  m1 := move{Right, Left}
  m2 := move{Right, Right}

  addParentChildEdges(n1, n2, m1) 
  addParentChildEdges(n1, n3, m2) 

  addParentChildEdges(n2, n4, m1)
  addParentChildEdges(n3, n4, m1)

  addParentChildEdges(n4, n5, m1)
  addParentChildEdges(n4, n6, m2)

  // Set scores:
  n6.score, n6.isScored = 0, true // Incorrect, should be 1
  n5.score, n5.isScored = 0.5, true // Incorrect, should be 0
  n4.score, n4.isScored = 0, true // Should be 1
  n3.score, n3.isScored = 0, true // Should be 1
  n2.score, n2.isScored = 0, true // Should be 1
  n1.score, n1.isScored = 0, true // Should be 1

  if err := solidifyScores(n1, 5); err != nil {
    t.Fatal(err.Error())
  }

  if n1.score != 1.0 {
    t.Fatal("Score update did not propagate.")
  }
  if n5.score != 0 {
    t.Fatal("Score update did not propagate.")
  }

  fmt.Println("finished TestSolidifyScore")
}
