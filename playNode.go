package main

import (
  "fmt"
  "strings"
  "errors"
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
// ==== End Move ====

// ==== playNode ==== 

// want: tree of optimal moves given the current move
type playNode struct {
  gs *gameState
  score float32 // +1 means Player1 wins, -1 means Player2 wins
  // Children nodes
  nextNodes map[move]*playNode
  // Parent nodes
  prevNodes map[move]*playNode
  // Whether or not the score of this node has been computed. Needed for score propagation
  isScored bool
}

// Construction
// ALWAYS copies the gamestate (I think??)
func createPlayNodeCopyGs(gs *gameState) *playNode {
  node := &playNode{gs.copyAndNormalize(), 0, make(map[move]*playNode), make(map[move]*playNode), false} 
  return node
}

// REUSES the gamestate, AND MUTATES THE ARGUMENT (I think??)
func createPlayNodeReuseGs(gs *gameState) *playNode {
  node := &playNode{gs, 0, make(map[move]*playNode), make(map[move]*playNode), false} 
  // MUTATES THE ARGUMENT
  node.gs.normalize()
  return node
}

// Scores
func (node *playNode) computeScore() float, err {
  // If the node is a leaf: 
  if len(node.nextNodes) == 0 {
    // Determine the score directly
    return node.getHeuristicScore(), nil
  } else {
    // Compute the score based on child moves. 
    _, score, err := node.getBestMoveAndScore()
    if err != nil {
      return 0, err
    }
    return score, nil
  }
}

func (node *playNode) updateScore() err {
  if score, err := node.computeScore(); err != nil {
    return err
  } else {
    node.score = score
    node.isScored = true
    return nil
  }
}

func (node *playNode) allChildrenAreScored() bool {
  for _, child := range node.nextNodes {
    if !child.isScored {
      return false
    }
  }
  return true
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

// Terminal states have one of the players eliminated; the game is over and no new moves are possible.
func (node *playNode) isTerminal() bool {
  return node.gs.player1.isEliminated() || node.gs.player2.isEliminated()
}

func (node *playNode) toString() string {
  return node.toStringImpl(0, 0, make(map[gameState]bool))
}

func (node *playNode) toTreeString(maxDepth int) string {
  return node.toStringImpl(0, maxDepth, make(map[gameState]bool))
}

func (node *playNode) toStringImpl(curDepth int, maxDepth int, printedStates map[gameState]bool) string {
  node.validateEdges(false)
  var sb strings.Builder
  buf := strings.Repeat(" ", curDepth)
  sb.WriteString(buf)
  sb.WriteString(fmt.Sprintf("playNode{gs:%s score:%f prevNodes:%+v ", node.gs.toString(), node.score, node.prevNodes)) 
  printedStates[*node.gs] = true
  if len(node.nextNodes) == 0 {
    sb.WriteString("leafNode}")
  } else {
    sb.WriteString("nextNodes:\n")
    for nextMove, nextNode := range node.nextNodes { 
      // One more space
      sb.WriteString(buf + " ")
      sb.WriteString(fmt.Sprintf("%+v\n", nextMove))
      if printedStates[*nextNode.gs] {
        sb.WriteString(buf + " ")
        sb.WriteString("<previously printed>")
      } else if curDepth < maxDepth {
        sb.WriteString(nextNode.toStringImpl(curDepth + 1, maxDepth, printedStates))
      } else {
        sb.WriteString(buf + " ")
        sb.WriteString("...")
      }
      sb.WriteString("\n")
    } 
    sb.WriteString(buf + "}")
  }
  return sb.String()
}


// Returns an error if any parent/child edge in the graph containing this node is not bi-directional
func (node *playNode) validateEdges(recurse bool) error {
  if recurse {
    return node.validateEdgesImpl(true, make(map[gameState]bool))
  } else {
    // Avoid memory allocation if it's unnecessary
    return node.validateEdgesImpl(false, nil)
  }
}

func (node *playNode) validateEdgesImpl(recurse bool, validatedStates map[gameState]bool) error {
  if validatedStates[*node.gs] {
    // We've visited this node before, return.
    return nil
  }

  // All children of this node must list this node as a parent.
  for childMove, childNode := range node.nextNodes {
    if childNode.prevNodes[childMove] != node {
      return errors.New(fmt.Sprintf("Child node does not contain parent that points to it: parent: %s, child %s", node.toTreeString(1), childNode.toString()))
    }
    // Recurse down the graph
    if recurse {
      if err := childNode.validateEdgesImpl(recurse, validatedStates); err != nil {
        return err
      }
    }
  }

  // All parents of this node must list this node as a child
  validatedStates[*node.gs] = true
  for parentMove, parentNode := range node.prevNodes {
    if parentNode.nextNodes[parentMove] != node {
      return errors.New(fmt.Sprintf("Parent node does not contain child that points to it: parent: %s, child %s",parentNode.toTreeString(1), node.toString()))
    }
    // Recurse up the graph 
    if recurse {
      if err := parentNode.validateEdgesImpl(recurse, validatedStates); err != nil {
        return err
      }
    }
  }

  return nil
}

// ==== end playNode ==== 