package main

import (
    "testing"
)

func TestSolveTreeValid(t *testing.T) {
  startState := gameState{
    player{2, 1}, player{2, 1}, Player1,
  }
  stateNode, visitedStates, _, solveErr := solve(&startState)
  gps := createGamePlayState(&startState)
  if solveErr != nil {
    t.Fatal(solveErr.Error())
  } 
  validateSolveNode(gps, stateNode, visitedStates, t)
}

func validateSolveNode(gps *gamePlayState, node *playNode, visitedStates map[gameState]*playNode, t *testing.T) {
  // Test one: our game play state should be valid
  if err := gps.validate(); err != nil {
    t.Fatalf("Game play state is invalid: %s", gps.toString())
  }
  // Test two: normalized state should be the same as the playNode state 
  if !gps.normalizedState.equals(node.gs) {
    t.Fatalf("Normalized play state does not match node state: play state: %+v, node state: %+v", *gps.normalizedState, *node.gs)
  }
  // For each possilbe move in the node:
  for nextMove, nextNode := range node.nextNodes {
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

    // Check that our playState is in our visited states map
    if visitedStates[*playState] == nil {
      t.Fatalf("Game state not found in visited states map: %+v", *playState)
    }

    // Apply the normalized move to our gps and recurse
    nextGps := gps.deepCopy()
    nextGps.playNormalizedTurn(nextMove)
    validateSolveNode(nextGps, nextNode, visitedStates, t)
  }
}

