package main

import (
  "fmt"
  "strings"
  "errors"
)

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

// ALWAYS use the left hand if both hands are identical
func normalizeMove(m move, gs *gameState) move {
  return move{normalizeHandForPlayer(m.playHand, &gs.player1), normalizeHandForPlayer(m.receiveHand, &gs.player2)}
}

// Like a game state but we maintain that the player hands are sorted
func (gs *gameState) normalize() *gameState {
  gs.player1.normalize() 
  gs.player2.normalize() 
  return gs
}

// Assume: computer is always player 1. Change later?
// 
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
  return node.toStringImpl(0)
}

func (node *playNode) toStringImpl(depth int) string {
  var sb strings.Builder
  buf := strings.Repeat(" ", depth)
  sb.WriteString(buf)
  sb.WriteString(fmt.Sprintf("playNode{gs:%s score:%f ",node.gs.toString(), node.score)) 
  if len(node.nextNodes) == 0 {
    sb.WriteString("leafNode:\n")
    sb.WriteString(node.gs.prettyString())
  } else {
    sb.WriteString("nextNodes:\n")
    for nextMove, nextNode := range node.nextNodes { 
      sb.WriteString(buf)
      // One more space
      sb.WriteString(fmt.Sprintf(" %+v\n", nextMove))
      sb.WriteString(nextNode.toStringImpl(depth + 1))
      sb.WriteString("\n")
    } 
  }
  sb.WriteString(buf + "}")
  return sb.String()
}

var DISTINCT_HANDS []Hand = []Hand{Left, Right}
// WLOG if the player can use (i.e. play or receive) either hand, he always uses his left hand.
var EITHER_HAND []Hand = []Hand{Left}

func getPossibleHands(p *player) []Hand {
  if p.lh == p.rh {
    return EITHER_HAND
  } else {
    return DISTINCT_HANDS
  }
}


func solve(gs *gameState) *playNode {
  result, err := solveDfs(createPlayNodeCopyGs(gs), make(map[gameState]*playNode, 5), 0)
  if err != nil {
    fmt.Println("Got error: " + err.Error())
  }
  // fmt.Println(result.toString())
  return result
}

func (node *playNode) getBestMoveAndScore() (*move, float32, error) {
  // Our best move is the move that puts our opponent in the worst position.
  // The score of the current node is the negative of the score of our opponent in the node after our best move.
  var worstNextScoreForOpp float32 = 2 // This is an impossible score, so we should always trigger an update in the loop.
  var bestMoveForUs *move = nil
  // log := len(node.nextNodes) > 1
  log := false 
  if log {
    fmt.Printf("-- Running getBestMoveAndScore() for %+v\n", node.gs)
  }
  for nextMove, nextNode := range node.nextNodes {
    oppScore := nextNode.scoreForCurrentPlayer() 
    if log {
      fmt.Printf("-- Node %+v, oppScore (for them): %f\n", nextNode.gs, oppScore)
    }
    if oppScore < worstNextScoreForOpp {
      worstNextScoreForOpp = oppScore
      bestMoveForUs = &nextMove
    }
  }
  if bestMoveForUs == nil || worstNextScoreForOpp > 1 || worstNextScoreForOpp < -1 {
    return nil, 0, errors.New(fmt.Sprintf("getBestMoveAndScore: play node is invalid: %+v", node))
  } else {
    // Note the negative sign!! worst score for opp is the best score for us.
    if log {
      fmt.Printf("-- result %+v, %f\n", bestMoveForUs, -worstNextScoreForOpp)
    }
    return bestMoveForUs, -worstNextScoreForOpp, nil
  }
}

const MAX_DEPTH int = 9

// Global map for saving pointers to the frontier 
// TODO: actually use?
// const frontier := map[gameState]*playNode

// Problems with this:
// how do you detect and kill loops? - need to memoize game states.
// there are redundant moves: [1,1],[1,1] should have one move, not four
//
func solveDfs(curNode *playNode, visitedStates map[gameState]*playNode, depth int) (*playNode, error) {
  // First: check if we've visited this state before. If so, 
  // abort the recursion as we're in a loop. TODO: also allow picking up where
  // we left off??
  curGs := *curNode.gs
  if visitedStates[curGs] != nil {
    fmt.Printf("-- Returning after detecting a loop. State %+v, score: %f\n", curNode.gs, curNode.score)
    // We're in a loop, abort recursion.
    return curNode, nil 
  }

  // Memoize the current node now so we can catch loops in recursive calls...
  // TODO: is this problematic for score evaluation?
  visitedStates[curGs] = curNode

  // If the game is over, determine the score and return
  if (curNode.gs.player1.isEliminated()) {
    curNode.score = -1
    if curNode.gs.turn != Player1 {
      return curNode, errors.New("invalid state: player eliminated when it's not their turn")
    }
    fmt.Printf("-- Returning after eliminating player1. State %+v, score: %f\n", curNode.gs, curNode.score)
    return curNode, nil
  } else if (curNode.gs.player2.isEliminated()) {
    curNode.score = 1
    if curNode.gs.turn != Player2 {
      return curNode, errors.New("invalid state: player eliminated when it's not their turn")
    }
    fmt.Printf("-- Returning after eliminating player2. State %+v, score: %f\n", curNode.gs, curNode.score)
    return curNode, nil
  }

  if depth >= MAX_DEPTH {
    // TODO: update frontier, set heuristic score, etc.
    // Let's set a heuristic score here: +/- 0.5 for every hand you/your opponent is missing
    curNode.score = curNode.getHeuristicScore()
    fmt.Printf("-- Returning after reaching maxDepth. State %+v, score: %f\n", curNode.gs, curNode.score)
    return curNode, nil 
  }

  // Recursively check all legal moves.
  // for playerHand
  player := curNode.gs.getPlayer()
  receiver := curNode.gs.getReceiver()


  for _, playerHand := range getPossibleHands(player) {
    if player.getHand(playerHand) == 0 {
      continue
    }
    for _, receiverHand := range getPossibleHands(receiver) {
      if receiver.getHand(receiverHand) == 0 {
        continue
      }
      curMove := move{playerHand, receiverHand}
      // A move is distinct if it leads to a distinct game state
      var nextNode *playNode

      // Check whether we've visited this node before
      // TODO: trickiness around distinct moves here? what if we've visited the state but via a different move before?
      if curNode.nextNodes[curMove] != nil {
        nextNode = curNode.nextNodes[curMove]
      } else {
        // Make sure the gamestate gets copied....
        nextState, _ := copyAndPlayTurn(curNode.gs, playerHand, receiverHand)

        // Add new state to cur state's children
        nextNode = createPlayNodeReuseGs(nextState)
        curNode.nextNodes[curMove] = nextNode
      }
      // Recurse, and bubble up errors
      _, err := solveDfs(nextNode, visitedStates, depth + 1)
      if err != nil {
        return curNode, err
      }
    }
  }


  // We're done evaluating all children. Determine our score
  _, curPlayerScore, err := curNode.getBestMoveAndScore()
  if err != nil {
    return curNode, err
  }
  // Note: curPlayerScore reflects score according to the current player, the absolute score needs to be adjusted (negative if player 2)
  curNode.score = scoreForPlayerToAbsoluteScore(curPlayerScore, curNode.gs.turn)
  return curNode, nil
}

