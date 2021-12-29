package main

import (
    "testing"
)

func TestSolveTreeValid(t *testing.T) {
  startState := gameState{
    player{2, 1}, player{2, 1}, Player1,
  }
  stateNode, visitedStates, solveErr := solve(&startState)
  if solveErr != nil {
    t.Fatal(solveErr.Error())
  } 
  validateSolveNode(&startState, stateNode, visitedStates, t)
}

func validateSolveNode(curState *gameState, node *playNode, visitedStates map[gameState]*playNode, t *testing.T) {
  normalizedState, swappedPlayer1, swappedPlayer2 := curState.copyAndNormalize()

  // Test one: normalized state should be the same as the playNode state 
  if !normalizedState.equals(node.gs) {
    t.Fatalf("Normalized play state does not match node state: play state: %+v, node state: %+v", *normalizedState, *node.gs)
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

    // Denormalize the move, apply it to the current state, and recurse.
    denormalizedMove := denormalizeMove(nextMove, swappedPlayer1, swappedPlayer2, curState.turn)
    nextStateDenormalized, err := curState.copyAndPlayTurn(denormalizedMove.playHand, denormalizedMove.receiveHand) 
    if err != nil {
      t.Fatal(err.Error())
    }
    validateSolveNode(nextStateDenormalized, nextNode, visitedStates, t)
  }
}

