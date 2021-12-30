package main

import (
  "fmt"
  "testing"
)

func TestSolveTreeValid(t *testing.T) {
  fmt.Println("starting TestSolveTreeValid")
  startState := gameState{
    player{2, 1}, player{2, 1}, Player1,
  }
  stateNode, existingStates, leaves, solveErr := solve(&startState)
  gps := createGamePlayState(&startState)
  if solveErr != nil {
    t.Fatal(solveErr.Error())
  } 
  validateSolveNode(gps, stateNode, make(map[gameState]bool, len(existingStates)), existingStates, leaves, t)
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
    bestMove, _, err := curNode.getBestMoveAndScore(false)
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
}