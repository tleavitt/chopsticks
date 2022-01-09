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
  // NOTE: this has to be a gameState map because there can be multiple distinct
  // parent states that lead to the same child state with the same move. Example:
  // {{3, 3}, {0, 2}, Player2} - {Left, Right} -> {{1, 3}, {0, 2} Player1}
  // {{1, 1}, {0, 2}, Player2} - {Left, Right} -> {{1, 3}, {0, 2} Player1}
  prevNodes map[gameState]*playNode
  // Whether or not the score of this node has been computed. Needed for score propagation
  isScored bool
  // Pointer to the loop node for this play node. Will be nil if not part of a loop.
  // TODO: should this be a global map instead?
  ln *loopNode
}

// Go needs generics dammit
func nodeMoveMapToString(nodeMap map[move]*playNode) string {
  var sb strings.Builder
  sb.WriteString("{")
  for m, n := range nodeMap {
    sb.WriteString(fmt.Sprintf("%+v:%s, ", m, n.toString()))
  }
  sb.WriteString("}")
  return sb.String()
}

func nodeStateMapToString(nodeMap map[gameState]*playNode) string {
  var sb strings.Builder
  sb.WriteString("{")
  for m, n := range nodeMap {
    sb.WriteString(fmt.Sprintf("%+v:%s, ", m, n.toString()))
  }
  sb.WriteString("}")
  return sb.String()
}

// Construction
// ALWAYS copies the gamestate (I think??)
func createPlayNodeCopyGs(gs *gameState) *playNode {
  node := &playNode{gs.copyAndNormalize(), 0, make(map[move]*playNode), make(map[gameState]*playNode), false, nil} 
  return node
}

// REUSES the gamestate, AND MUTATES THE ARGUMENT (I think??)
func createPlayNodeReuseGs(gs *gameState) *playNode {
  node := &playNode{gs, 0, make(map[move]*playNode), make(map[gameState]*playNode), false, nil} 
  // MUTATES THE ARGUMENT
  node.gs.normalize()
  return node
}

// Scores
// Note: the node must not be a leaf (i.e. it must have children) or this function will fail
func getBestMoveAndScore(childNodes map[move]*playNode, log bool, allowUnscoredChild bool) (move, float32, error) {
  // Our best move is the move that puts our opponent in the worst position.
  // The score of the current node is the negative of the score of our opponent in the node after our best move.
  var worstNextScoreForOpp float32 = 2 // This is an impossible score, so we should always trigger an update in the loop.
  var bestMoveForUs move // This should always get updated.

  for nextMove, nextNode := range childNodes {

    if !allowUnscoredChild && !nextNode.isScored {
      return bestMoveForUs, 0, errors.New(fmt.Sprintf("Child node is not scored: %s", nextNode.toString()))
    }
    oppScore := nextNode.scoreForCurrentPlayer() 
    if log {
      fmt.Printf("-- Move: %+v, GS %+v, oppScore (for them): %f, worstNextScoreForOpp: %f, bestMoveForUs: %+v\n", nextMove, nextNode.gs, oppScore, worstNextScoreForOpp, bestMoveForUs)
    }
    if oppScore < worstNextScoreForOpp {
      worstNextScoreForOpp = oppScore
      // Tricky bug! next move gets reused within the for loop, need to copy. Don't use pointers here.
      bestMoveForUs = nextMove
      if log {
        fmt.Printf("--- Update triggered, new worstNextScoreForOpp: %f, new bestMoveForUs %+v\n", worstNextScoreForOpp, bestMoveForUs)
      }
    }
  }
  if worstNextScoreForOpp > 1 || worstNextScoreForOpp < -1 {
    return bestMoveForUs, 0, errors.New(fmt.Sprintf("getBestMoveAndScore: no best move found, worst next score for opp: %f", worstNextScoreForOpp))
  } else {
    // Note the negative sign!! worst score for opp is the best score for us.
    if log {
      fmt.Printf("-- result %+v, %f\n", bestMoveForUs, -worstNextScoreForOpp)
    }
    return bestMoveForUs, -worstNextScoreForOpp, nil
  }
}

func (node *playNode) getBestMoveAndScore(log bool, allowUnscoredChild bool) (move, float32, error) {
  if log {
    fmt.Printf("-- Running getBestMoveAndScore() for %+v\n", node.gs)
  }
  return getBestMoveAndScore(node.nextNodes, log, allowUnscoredChild)
}


func (node *playNode) computeScore(allowUnscoredChild bool) (float32, error) {
  // If the node is a leaf: 
  if len(node.nextNodes) == 0 {
    // Determine the score directly
    return node.getHeuristicScore(), nil
  } else {
    // Compute the score based on child moves. 
    _, score, err := node.getBestMoveAndScore(false, allowUnscoredChild)
    if err != nil {
      return 0, err
    }
    return score, nil
  }
}

func (node *playNode) updateScore() error {
  if score, err := node.computeScore(false); err != nil {
    return err
  } else {
    node.score = score
    node.isScored = true
    if DEBUG {
      fmt.Println("Computed score for node: " + node.toString())
    }
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
  // Would be nice to do this here but it could cause problems when debugging
  // if node.validateEdges(false)
  var sb strings.Builder
  buf := strings.Repeat(" ", curDepth)
  sb.WriteString(buf)
  sb.WriteString(fmt.Sprintf("playNode{gs:%s score:%f, isScored:%t prevNodes:%+v ", node.gs.toString(), node.score, node.isScored, node.prevNodes)) 
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

// Returns the move that takes the map to the given node, or nil if the given node is not a value in the map.
func findNodeInMap(node *playNode, nodeMap map[move]*playNode) *move {
  for m, n := range nodeMap {
    if n == node {
      return &m
    }
  }
  return nil
}

func addParentChildEdges(parent *playNode, child *playNode, m move) {
  addParentEdge(parent, child, m)
  addChildEdge(parent, child)
}

func addParentEdge(parent *playNode, child *playNode, m move) {
  parent.nextNodes[m] = child 
}

func addChildEdge(parent *playNode, child *playNode) {
  child.prevNodes[*parent.gs] = parent
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

  if validatedStates != nil {
    validatedStates[*node.gs] = true
  }

  // All children of this node must list this node as a parent.
  for _, childNode := range node.nextNodes {
    if childNode.prevNodes[*node.gs] != node {
      return errors.New(fmt.Sprintf("Child node does not contain parent that points to it: parent: %s, child %s, child prev nodes: %+v", node.toTreeString(1), childNode.toString(), nodeStateMapToString(childNode.prevNodes)))
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
  for _, parentNode := range node.prevNodes {
    parentMove := findNodeInMap(node, parentNode.nextNodes)
    if parentMove == nil {
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