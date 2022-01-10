package main

import (
  "fmt"
  "testing"
)

func testExploreStates(numFingers int8, t *testing.T) {
  fmt.Println("starting TestExploreStates")
  prevNumFingers := setNumFingers(numFingers)
  startState := &gameState{
    player{1, 1}, player{1, 1}, Player1,
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
    for _, node := range loop {
      fmt.Printf("%s, ", node.gs.toString())
    }
    fmt.Printf("\n")
  }
  // Not necessarily the case with loops
  for _, leafNode := range leaves {
    if len(leafNode.nextNodes) > 0 {
      t.Fatalf("Leaf node has children: %s", leafNode.toString())
    }
  }

  // Yay it's fixed now
  if err := startNode.validateEdges(true); err != nil {
    t.Fatal(err)
  }

  // fmt.Println(startNode.toTreeString(15))
  setNumFingers(prevNumFingers)
  fmt.Println("finished TestExploreStates")
}

func TestExploreStates3(t *testing.T) {
  testExploreStates(3, t)
}

func TestExploreStates4(t *testing.T) {
  testExploreStates(4, t)
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
    for _, node := range loop {
      fmt.Printf("%s, ", node.gs.toString())
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
  if err := startNode.validateEdges(true); err != nil {
    t.Fatal(err)
  }

  // fmt.Println(startNode.toTreeString(15))
  setNumFingers(prevNumFingers)
  fmt.Println("finished TestExploreLoop")
}

func expectInvalidGraph(startNode *playNode, t *testing.T) {
  var err error
  if err = startNode.validateEdges(true); err == nil {
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
