package main

// import (
//   "fmt"
//   "github.com/pkg/errors"
// )

type move struct {
  playHand Hand
  receiveHand Hand
}

// Assume: computer is always player 1. Change later?
// 
// want: tree of optimal moves given the current move
type playNode struct {
  gs gameState
  score float32
  nextStates map[move]playNode
}

// Global map for easier lookups?
// const stateMap := map[gameState]*playNode

func solveDfs(rootNode *playNode) *playNode {
  // If the game is over, determine the score and return
  rootNode.gs = gs
  if (rootNode.gs.player1.isEliminated()) {
    rootNode.score = -1
    return rootNode
  } else if (rootNode.gs.player2.isEliminated()) {
    rootNode.score = 1
    return rootNode
  }

  // Recursively check all legal moves.
  // for playerHand
  // TODO
  return rootNode
}

