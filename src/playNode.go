package main

import (
  "fmt"
  "strings"
  "errors"
  "math"
)

// ==== Move ==== 
type Move struct {
  PlayerHand Hand
  ReceiverHand Hand
}

func (m *Move) toString() string {
  return toString(m.PlayerHand) + " -> " + toString(m.ReceiverHand)
}

func normalizeHandForPlayer(h Hand, p *Player) Hand {
  if p.isEliminated() {
    fmt.Println("Warning: normalizing hand for eliminated Player")
  }
  if p.Lh == p.Rh {
    return Left
  } else {
    return h
  }
}
// ==== End Move ====

// ==== PlayNode ==== 

// want: tree of optimal Moves given the current Move
type PlayNode struct {
  gs *GameState
  score float32 // +1 means Player1 wins, -1 means Player2 wins
  // Children nodes
  nextNodes map[Move]*PlayNode
  // Parent nodes
  // NOTE: this has to be a GameState map because there can be multiple distinct
  // parent states that lead to the same child state with the same Move. Example:
  // {{3, 3}, {0, 2}, Player2} - {Left, Right} -> {{1, 3}, {0, 2} Player1}
  // {{1, 1}, {0, 2}, Player2} - {Left, Right} -> {{1, 3}, {0, 2} Player1}
  prevNodes map[GameState]*PlayNode
  // Whether or not the score of this node has been computed. Needed for score propagation
  isScored bool
  // Pointer to the loop node(s) for this play node. Will be empty if not part of a loop.
  // TODO: should this be a global map instead?
  lns []*loopNode
}

// Go needs generics dammit
func nodeMoveMapToString(nodeMap map[Move]*PlayNode) string {
  var sb strings.Builder
  sb.WriteString("{")
  for m, n := range nodeMap {
    sb.WriteString(fmt.Sprintf("%+v:%s, ", m, n.toString()))
  }
  sb.WriteString("}")
  return sb.String()
}

func nodeStateMapToString(nodeMap map[GameState]*PlayNode) string {
  var sb strings.Builder
  sb.WriteString("{")
  for m, n := range nodeMap {
    sb.WriteString(fmt.Sprintf("%+v:%s, ", m, n.toString()))
  }
  sb.WriteString("}")
  return sb.String()
}

// Construction
// ALWAYS copies the gamestate
func createPlayNodeCopyGs(gs *GameState) *PlayNode {
  node := &PlayNode{gs.copyAndNormalize(), 0, make(map[Move]*PlayNode), make(map[GameState]*PlayNode), false, []*loopNode{}} 
  return node
}

// REUSES the gamestate, AND MUTATES THE ARGUMENT
func createPlayNodeReuseGs(gs *GameState) *PlayNode {
  node := &PlayNode{gs, 0, make(map[Move]*PlayNode), make(map[GameState]*PlayNode), false, []*loopNode{}} 
  // MUTATES THE ARGUMENT
  node.gs.normalize()
  return node
}

// Scores
// Note: the node must not be a leaf (i.e. it must have children) or this function will fail
func getBestMoveAndScoreForCurrentPlayer(childNodes map[Move]*PlayNode, log bool, allowUnscoredChild bool) (Move, float32, error) {
  // Our best Move is the Move that puts our opponent in the worst position.
  // The score of the current node is the negative of the score of our opponent in the node after our best Move.
  var worstNextScoreForOpp float32 = 2 // This is an impossible score, so we should always trigger an update in the loop.
  var bestMoveForUs Move // This should always get updated.

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
      // Tricky bug! next Move gets reused within the for loop, need to copy. Don't use pointers here.
      bestMoveForUs = nextMove
      if log {
        fmt.Printf("--- Update triggered, new worstNextScoreForOpp: %f, new bestMoveForUs %+v\n", worstNextScoreForOpp, bestMoveForUs)
      }
    }
  }
  if worstNextScoreForOpp > 1 || worstNextScoreForOpp < -1 {
    return bestMoveForUs, 0, errors.New(fmt.Sprintf("getBestMoveAndScoreForCurrentPlayer: no best Move found, worst next score for opp: %f", worstNextScoreForOpp))
  } else {
    // Note the negative sign!! worst score for opp is the best score for us.
    if log {
      fmt.Printf("-- result %+v, %f\n", bestMoveForUs, -worstNextScoreForOpp)
    }
    return bestMoveForUs, -worstNextScoreForOpp, nil
  }
}

func (node *PlayNode) getBestMoveAndScoreForCurrentPlayer(log bool, allowUnscoredChild bool) (Move, float32, error) {
  if log {
    fmt.Printf("-- Running getBestMoveAndScoreForCurrentPlayer() for %+v\n", node.gs)
  }
  return getBestMoveAndScoreForCurrentPlayer(node.nextNodes, log, allowUnscoredChild)
}


func (node *PlayNode) computeScore(allowUnscoredChild bool) (float32, error) {
  // If all children are scored, return the best score based on the children.
  if len(node.nextNodes) == 0 {
    // Determine the score directly
    return node.getHeuristicScore(), nil
  } else {
    // Compute the score based on child Moves. 
    _, scoreForCurrentPlayer, err := node.getBestMoveAndScoreForCurrentPlayer(false, allowUnscoredChild)
    if err != nil {
      return 0, err
    }
    return turnToSign(node.gs.T) * scoreForCurrentPlayer, nil
  }
}

func (node *PlayNode) updateScore() error {
  if score, err := node.computeScore(false); err != nil {
    if DEBUG {
      fmt.Println("ERR when updating score: " + node.toString())
    }
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

func (node *PlayNode) allChildrenAreScored() bool {
  for _, child := range node.nextNodes {
    if !child.isScored {
      return false
    }
  }
  return true
}

func getHeuristicScoreForPlayer(p *Player) float32 {
  if p.Lh == 0 {
    if p.Rh == 0 {
      return -1
    } else {
      return -0.5
    }
  } else {
    return 0
  }
}

// TODO: aggressive/defensive, apply more/less weight to my score vs their score
func (node *PlayNode) getHeuristicScore() float32 {
  p1Heuristic := getHeuristicScoreForPlayer(&node.gs.Player1)
  p2Heuristic := getHeuristicScoreForPlayer(&node.gs.Player2)

  if p1Heuristic == -1 {
    if p2Heuristic == -1 {
      // This is an invalid state where both Players are eliminated, but we don't have to modify the heuristics. just return 0
      return 0
    } else {
      // p2 wins, return -1
      return -1
    }
  } else {
    if p2Heuristic == -1 {
      // p1 wins, return +1
      return 1
    } else {
      // Don't modify the heuristics at all, we're not in a terminal case.
      return p1Heuristic - p2Heuristic
    }
  }
}

func (node *PlayNode) getHeuristicScoreForCurrentPlayer() float32 {
  return turnToSign(node.gs.T) * node.getHeuristicScore()
}

func turnToSign(t Turn) float32 {
  if t == Player1 {
    return 1
  } else {
    return -1
  }
}

// Flips the sign for Player 2 scores
func scoreForPlayerToAbsoluteScore(score float32, turn Turn) float32 {
  return turnToSign(turn) * score
}

// For this function, +1 means the current Player (i.e. the Player whose turn it is) is winning, -1 means losing.
func (node *PlayNode) scoreForCurrentPlayer() float32 {
  return turnToSign(node.gs.T) * node.score
}

// Terminal states have one of the Players eliminated; the game is over and no new Moves are possible.
func (node *PlayNode) isTerminal() bool {
  return node.gs.Player1.isEliminated() || node.gs.Player2.isEliminated()
}

func (node *PlayNode) toString() string {
  return node.toStringImpl(0, 0, make(map[GameState]bool))
}

func (node *PlayNode) toTreeString(maxDepth int) string {
  return node.toStringImpl(0, maxDepth, make(map[GameState]bool))
}

func (node *PlayNode) toStringImpl(curDepth int, maxDepth int, printedStates map[GameState]bool) string {
  // Would be nice to do this here but it could cause problems when debugging
  // if node.validateEdges(false)
  var sb strings.Builder
  buf := strings.Repeat(" ", curDepth)
  sb.WriteString(buf)
  sb.WriteString(fmt.Sprintf("PlayNode{gs:%s score:%f, isScored:%t prevNodes:%+v lns: %v ", node.gs.toString(), node.score, node.isScored, node.prevNodes, node.lns)) 
  printedStates[*node.gs] = true
  if len(node.nextNodes) == 0 {
    sb.WriteString("leafNode}")
  } else {
    sb.WriteString("nextNodes:\n")
    for _, nextNode := range node.nextNodes { 
      // One more space
      if printedStates[*nextNode.gs] {
        sb.WriteString(buf + " ")
        sb.WriteString(fmt.Sprintf("%p <previously printed>", nextNode))
      } else if curDepth < maxDepth {
        sb.WriteString(nextNode.toStringImpl(curDepth + 1, maxDepth, printedStates))
      } else {
        sb.WriteString(buf + " ")
        sb.WriteString(fmt.Sprintf("%p", nextNode))
      }
      sb.WriteString("\n")
    } 
    sb.WriteString(buf + "}")
  }
  return sb.String()
}

// Returns the Move that takes the map to the given node, or nil if the given node is not a value in the map.
func findNodeInMap(node *PlayNode, nodeMap map[Move]*PlayNode) *Move {
  for m, n := range nodeMap {
    if n == node {
      return &m
    }
  }
  return nil
}

func addParentChildEdges(parent *PlayNode, child *PlayNode, m Move) {
  addParentEdge(parent, child, m)
  addChildEdge(parent, child)
}

func addParentEdge(parent *PlayNode, child *PlayNode, m Move) {
  parent.nextNodes[m] = child 
}

func addChildEdge(parent *PlayNode, child *PlayNode) {
  child.prevNodes[*parent.gs] = parent
}

// Returns an error if any parent/child edge in the graph containing this node is not bi-directional
func (node *PlayNode) validateEdges(recurse bool) (int, int, error) {
  if recurse {
    return node.validateEdgesImpl(true, []*PlayNode{node}, make(map[GameState]bool))
  } else {
    // Avoid memory allocation if it's unnecessary
    return node.validateEdgesImpl(false, nil, nil)
  }
}

func (node *PlayNode) validateEdgesImpl(recurse bool, curPath []*PlayNode, validatedStates map[GameState]bool) (int, int, error) {
  // curDepth = current depth of this node
  // minDepth: minimum depth of children of this node
  // maxDepth: maximum depth of children of this node.
  curDepth := len(curPath)
  if validatedStates[*node.gs] {
    // We've visited this node before, return.
    // EDIT: is this an error?
    return curDepth, curDepth, nil
  }

  // Is there a bug in this??
  if validatedStates != nil {
    validatedStates[*node.gs] = true
  }

  // All children of this node must list this node as a parent.
  minChildDepth, maxChildDepth := math.MaxInt32, 0
  for _, childNode := range node.nextNodes {
    if childNode.prevNodes[*node.gs] != node {
      return curDepth, curDepth, errors.New(fmt.Sprintf("Child node does not contain parent that points to it: parent: %s, child %s, child prev nodes: %+v", node.toTreeString(1), childNode.toString(), nodeStateMapToString(childNode.prevNodes)))
    }
    if recurse {
      nextPath := append(curPath, childNode)
      min, max, err := childNode.validateEdgesImpl(recurse, nextPath, validatedStates)
      if err != nil {
        return curDepth, curDepth, err
      }
      if min < minChildDepth {
        minChildDepth = min
      }
      if max > maxChildDepth {
        maxChildDepth = max
      }
    }
  }
  // If we updated in the loop, change our min/max depth values. Otherwise keep them as curDepth. That way leaves
  // always return curDepth for both.
  minDepth, maxDepth := curDepth, curDepth
  if maxChildDepth > 0 {
    minDepth = minChildDepth
    maxDepth = maxChildDepth
  }

  // All parents of this node must list this node as a child
  validatedStates[*node.gs] = true
  for _, parentNode := range node.prevNodes {
    parentMove := findNodeInMap(node, parentNode.nextNodes)
    if parentMove == nil {
       return curDepth, curDepth, errors.New(fmt.Sprintf("Parent node does not contain child that points to it: parent: %s, child %s",parentNode.toTreeString(1), node.toString()))
    }
    // Recurse up the graph to catch invalid parents
    if recurse {
      if _, _, err := parentNode.validateEdgesImpl(recurse, curPath[:len(curPath) - 1], validatedStates); err != nil {
        return curDepth, curDepth, err
      }
    }
  }

  return minDepth, maxDepth, nil
}

// ==== end PlayNode ==== 