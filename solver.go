package main

// import (
//   "fmt"
//   "github.com/pkg/errors"
// )

type move struct {
  playHand Hand
  receiveHand Hand
}

// want: tree of optimal moves given the current move

type playNode struct {
  gs gameState
  score float32
  connections map[move]playNode
  isMyTurn bool
}

