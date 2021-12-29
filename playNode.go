package main

import (
  "fmt"
  "strings"
)

// ==== Move ==== 
type move struct {
  playHand Hand
  receiveHand Hand
}

func (m *move) toString() string {
  return toString(m.playHand) + " -> " + toString(m.receiveHand)
}

func normalizeHandForPlayer(h Hand, p *player) Hand {
  if p.isEliminated() {
    fmt.Println("Warning: normalizing hand for eliminated player")
  }
  if p.lh == p.rh {
    return Left
  } else {
    return h
  }
}

// ALWAYS use the left hand if both hands are identical, for either player
func normalizeMove(m move, gs *gameState) move {
  return move{normalizeHandForPlayer(m.playHand, &gs.player1), normalizeHandForPlayer(m.receiveHand, &gs.player2)}
}
// ==== End Move ====

// ==== playNode ==== 

// want: tree of optimal moves given the current move
type playNode struct {
  gs *gameState
  score float32 // +1 means Player1 wins, -1 means Player2 wins
  nextNodes map[move]*playNode
}

func getHeuristicScoreForPlayer(p *player) float32 {
  if p.lh == 0 {
    if p.rh == 0 {
      return -1
    } else {
      return -0.5
    }
  } else {
    return 0
  }
}

// TODO: aggressive/defensive, apply more/less weight to my score vs their score
func (node *playNode) getHeuristicScore() float32 {
  var score float32 = 0
  // +1 means player 1 wins, -1 means player 2 wins. SO: we Need a negative sign in front of the second player
  score += getHeuristicScoreForPlayer(&node.gs.player1)
  score -= getHeuristicScoreForPlayer(&node.gs.player2)
  return score
}

func turnToSign(t Turn) float32 {
  if t == Player1 {
    return 1
  } else {
    return -1
  }
}

// Flips the sign for player 2 scores
func scoreForPlayerToAbsoluteScore(score float32, turn Turn) float32 {
  return turnToSign(turn) * score
}

// For this function, +1 means the current player (i.e. the player whose turn it is) is winning, -1 means losing.
func (node *playNode) scoreForCurrentPlayer() float32 {
  return turnToSign(node.gs.turn) * node.score
}

// ALWAYS copies the gamestate (I think??)
func createPlayNodeCopyGs(gs *gameState) *playNode {
  copyGs := *gs
  node := &playNode{&copyGs, 0, make(map[move]*playNode)} 
  node.gs.normalize()
  return node
}

// REUSES the gamestate, AND MUTATES THE ARGUMENT (I think??)
func createPlayNodeReuseGs(gs *gameState) *playNode {
  node := &playNode{gs, 0, make(map[move]*playNode)} 
  // MUTATES THE ARGUMENT
  node.gs.normalize()
  return node
}

func (node *playNode) toString() string {
  return node.toStringImpl(0, 0, make(map[gameState]bool))
}

func (node *playNode) toTreeString(maxDepth int) string {
  return node.toStringImpl(0, maxDepth, make(map[gameState]bool))
}

func (node *playNode) toStringImpl(curDepth int, maxDepth int, printedStates map[gameState]bool) string {
  var sb strings.Builder
  buf := strings.Repeat(" ", curDepth)
  sb.WriteString(buf)
  sb.WriteString(fmt.Sprintf("playNode{gs:%s score:%f ", node.gs.toString(), node.score)) 
  printedStates[node.gs] = true
  if len(node.nextNodes) == 0 {
    sb.WriteString("leafNode:\n")
    sb.WriteString(node.gs.prettyString())
  } else {
    sb.WriteString("nextNodes:\n")
    for nextMove, nextNode := range node.nextNodes { 
      sb.WriteString(buf)
      // One more space
      sb.WriteString(fmt.Sprintf(" %+v\n", nextMove))
      if printedStates[*nextNode.gs] != nil {
        sb.WriteString("<previously printed>")
      } else if curDepth < maxDepth {
        sb.WriteString(nextNode.toStringImpl(curDepth + 1, maxDepth, visitedStates))
      } else {
        sb.WriteString("...")
      }
      sb.WriteString("\n")
    } 
  }
  sb.WriteString(buf + "}")
  return sb.String()
}

// ==== end playNode ==== 