package main

import (
  "fmt"
  "strings"
  // "github.com/pkg/errors"
)

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
  nextNodes map[move]*playNode
}

func (node *playNode) toString() string {
  var sb strings.Builder
  sb.WriteString(fmt.Sprintf("playNode{gs:%s score:%f nextNodes:\n",node.gs.toString(), node.score)) 
  // No idea what this will do
  // return fmt.Sprintf("%+v\n", node)
  return sb.String()
}

var HANDS []Hand = []Hand{Left, Right}


func solve(gs *gameState) *playNode {
  result := solveDfs(&playNode{*gs, 0, make(map[move]*playNode)}, 0)
  fmt.Println(result.toString())
  return result
}

// Global map for easier lookups?
// const stateMap := map[gameState]*playNode

func solveDfs(curNode *playNode, depth int) *playNode {
  // If the game is over, determine the score and return
  if (curNode.gs.player1.isEliminated()) {
    curNode.score = -1
    return curNode
  } else if (curNode.gs.player2.isEliminated()) {
    curNode.score = 1
    return curNode
  }

  if depth >= 5 {
    return curNode 
  }

  // Recursively check all legal moves.
  // for playerHand
  // TODO

  player := curNode.gs.getPlayer()
  receiver := curNode.gs.getReceiver()
  for _, playerHand := range HANDS {
    if player.getHand(playerHand) == 0 {
      continue
    }
    for _, receiverHand := range HANDS {
      if receiver.getHand(receiverHand) == 0 {
        continue
      }
      curMove := move{playerHand, receiverHand}
      // Make sure the gamestate gets copied....
      nextState, _ := copyAndPlayTurn(curNode.gs, playerHand, receiverHand)

      // Add new state to cur state's children
      nextNode := playNode{*nextState, 0, make(map[move]*playNode)}
      curNode.nextNodes[curMove] = &nextNode
      // Recurse
      solveDfs(&nextNode, depth + 1)
    }
  }
  return curNode
}

