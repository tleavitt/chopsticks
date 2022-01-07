package main

import (
  "fmt"
  "testing"
  "errors"
)

func TestSolveTreeValid(t *testing.T) {
  fmt.Println("starting TestSolveTreeValid")
  prevNumFingers := setNumFingers(4)
  startState := gameState{
    player{2, 1}, player{2, 1}, Player1,
  }
  stateNode, existingStates, leaves, solveErr := solve(&startState)
  gps := createGamePlayState(&startState)
  if solveErr != nil {
    t.Fatal(solveErr.Error())
  } 
  validateSolveNode(gps, stateNode, make(map[gameState]bool, len(existingStates)), existingStates, leaves, t)
  setNumFingers(prevNumFingers)
  fmt.Println("finished TestSolveTreeValid")
}

func validateSolveNode(gps *gamePlayState, node *playNode, visitedStates map[gameState]bool, 
                       existingStates map[gameState]*playNode, leaves map[gameState]*playNode, t *testing.T) {

  // Test one: our game play state should be valid
  if err := gps.validate(); err != nil {
    t.Fatalf("Game play state is invalid: %s", gps.toString())
  }

  // Record that we've visited this state 
  visitedStates[*gps.normalizedState] = true

  // Check that our playState is somethihng we've seen when solving
  if existingStates[*gps.normalizedState] == nil {
    t.Fatalf("Game state not found in visited states map: %+v", *gps.normalizedState)
  }

  // Test two: normalized state should be the same as the playNode state 
  if !gps.normalizedState.equals(node.gs) {
    t.Fatalf("Normalized play state does not match node state: play state: %+v, node state: %+v", *gps.normalizedState, *node.gs)
  }

  // Test three: node should be scored
  if !node.isScored {
    t.Fatal("Node is unscored: " + node.toString())
  }
  // If this node has no children, then it should be a leaf:
  if len(node.nextNodes) == 0 {
    if leaves[*node.gs] == nil {
      t.Fatalf("Game state has no children but is not a leaf: %+v", *node.gs)
    }
  }

  // For each possilbe move in the node:
  for nextMove, nextNode := range node.nextNodes {
    // If we've visited this state before, continue
    if visitedStates[*nextNode.gs] {
      continue
    }

    // Check that applying the move to the current node state gives you the state in the next node
    playState, err := node.gs.copyAndPlayTurn(nextMove.playHand, nextMove.receiveHand) 
    if err != nil {
      t.Fatal(err.Error())
    }
    // Normalize in place
    playState.normalize()
    if !playState.equals(nextNode.gs) {
      t.Fatalf("Normalized play state does not match node state after move: play state: %+v, node state: %+v, move: %+v", *playState, *nextNode.gs, nextMove)
    }


    // Apply the normalized move to our gps and recurse
    nextGps := gps.deepCopy()
    nextGps.playNormalizedTurn(nextMove)
    validateSolveNode(nextGps, nextNode, visitedStates, existingStates, leaves, t)
  }
}

func TestSolveBestMoves(t *testing.T) {
  fmt.Println("starting TestSolveBestMoves")
  startState := gameState{
    player{1, 1}, player{1, 1}, Player1,
  }
  stateNode, _, _, err := solve(&startState)
  if err != nil {
    t.Fatal(err.Error())
  }

  var i int
  var curNode = stateNode
  var gameResult GameResult
  for i, gameResult = 0, checkGameResult(curNode.gs); gameResult == Ongoing; i, gameResult = i+1, checkGameResult(curNode.gs) {
    if len(curNode.nextNodes) == 0 {
      fmt.Println("Hit leaf node, exiting")
      break;
    }
    bestMove, _, err := curNode.getBestMoveAndScore(false, false)
    if err != nil {
      t.Fatal(err.Error())
    }
    node, ok := curNode.nextNodes[bestMove]
    if !ok {
      t.Fatalf("Best move not found in node states: %+v, %s", bestMove, curNode.toTreeString(1))
    }
    fmt.Printf("Previous node: %s, best move: %+v, next node: %s\n", curNode.toString(), bestMove, node.toString())
    curNode = node
  }
  if gameResult == Player1Wins {
    fmt.Println("Player 1 wins")
  } else if gameResult == Player2Wins {
    fmt.Println("Player 2 wins")
  } else {
    fmt.Println("Computer ran out of moves!")
  }
  fmt.Println("finished TestSolveBestMoves")
}

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

func TestInvalidGraphParent(t *testing.T) {
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

func TestInvalidGraphChild(t *testing.T) {
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

func ensureAllNodesScored(root *playNode) error {
  return ensureAllNodesScoredImpl(root, make(map[gameState]bool))
}

func ensureAllNodesScoredImpl(root *playNode, visitedStates map[gameState]bool) error {
  if visitedStates[*root.gs] {
    return nil
  }
  visitedStates[*root.gs] = true
  if !root.isScored {
    return errors.New("Found unscored node: " + root.toString())
  }
  for _, child := range root.nextNodes {
    if err := ensureAllNodesScoredImpl(child, visitedStates); err != nil {
      return err
    }
  }
  return nil
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

  addParentChildEdges(dadNode, dadNode, move{Right, Left})

  // Score
  leaves := make(map[gameState]*playNode, 1)
  leaves[*sonNode.gs] = sonNode
  if err := propagateScores(leaves, 5); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(grandpaNode)

  fmt.Println("starting TestPropagateScores")
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
  if err := propagateScores(leaves, 5); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(one)

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

  // What are the leaves here?
  leaves := make(map[gameState]*playNode, 1)
  leaves[*fourNode.gs] = fourNode
  // Should require exactly two nodes on the frontier (two and two prime)
  if err := propagateScores(leaves, 7); err != nil {
    t.Fatal(err.Error())
  }

  ensureAllNodesScored(entryNode)

  fmt.Println("starting TestPropagateScoresLoop")
}
