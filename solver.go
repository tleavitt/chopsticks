package main

import (
  "fmt"
  "errors"
)

var DISTINCT_HANDS []Hand = []Hand{Left, Right}
// WLOG if the player can use (i.e. play or receive) either hand, they always use their left hand.
var EITHER_HAND []Hand = []Hand{Left}

func getPossibleHands(p *player) []Hand {
  if p.lh == p.rh {
    return EITHER_HAND
  } else {
    return DISTINCT_HANDS
  }
}

func solve(gs *gameState) (*playNode, map[gameState]*playNode, error) {
  var visitedStates = make(map[gameState]*playNode, 5)
  result, err := solveDfs(createPlayNodeCopyGs(gs), visitedStates, 0)
  // fmt.Println(result.toString())
  fmt.Println("Generated move tree with " + fmt.Sprint(len(visitedStates)) + " nodes.")
  return result, visitedStates, err
}

func (node *playNode) getBestMoveAndScore(log bool) (move, float32, error) {
  // Our best move is the move that puts our opponent in the worst position.
  // The score of the current node is the negative of the score of our opponent in the node after our best move.
  var worstNextScoreForOpp float32 = 2 // This is an impossible score, so we should always trigger an update in the loop.
  var bestMoveForUs move // This should always get updated.
  if log {
    fmt.Printf("-- Running getBestMoveAndScore() for %+v\n", node.gs)
  }
  for nextMove, nextNode := range node.nextNodes {
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
    return bestMoveForUs, 0, errors.New(fmt.Sprintf("getBestMoveAndScore: play node is invalid: %+v", node))
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
    // We should catch loops before we make recursive calls, so error if we detect a loop
    return curNode, errors.New("detected loop at beginning of recursive call: " + curNode.toString())
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
        nextState, err := curNode.gs.copyAndPlayTurn(playerHand, receiverHand)
        if err != nil {
          return curNode, err
        }

        // HERE we have to check for possible loops, and if so, add the loop connection
        // in our graph

        existingNode, exists := visitedStates[*nextState]
        if exists {
          if DEBUG {
            fmt.Printf(fmt.Sprintf("Found loop in move tree, not recursing further. CurNode: %s, loop move: %+v\n", curNode.toString(), curMove))
          }
          curNode.nextNodes[curMove] = existingNode
          continue
        }

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
  _, curPlayerScore, err := curNode.getBestMoveAndScore(false)
  if err != nil {
    return curNode, err
  }
  // Note: curPlayerScore reflects score according to the current player, the absolute score needs to be adjusted (negative if player 2)
  curNode.score = scoreForPlayerToAbsoluteScore(curPlayerScore, curNode.gs.turn)
  return curNode, nil
}

