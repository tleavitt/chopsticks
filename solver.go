package main

import (
  "fmt"
  "errors"
)


const DEFAULT_MAX_DEPTH int = 15

func solve(gs *gameState) (*playNode, map[gameState]*playNode, map[gameState]*playNode, error) {
  visitedStates := make(map[gameState]*playNode, 5)
  leaves := make(map[gameState]*playNode, 5)
  result, err := solveDfs(createPlayNodeCopyGs(gs), visitedStates, leaves, 0, DEFAULT_MAX_DEPTH)
  // fmt.Println(result.toString())
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves), root score: %f", len(visitedStates), len(leaves), result.score))
  }
  return result, visitedStates, leaves, err
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
    return bestMoveForUs, 0, errors.New(fmt.Sprintf("getBestMoveAndScore: play node is invalid: %+s; worst next score for opp: %f", node.toString(), worstNextScoreForOpp))
  } else {
    // Note the negative sign!! worst score for opp is the best score for us.
    if log {
      fmt.Printf("-- result %+v, %f\n", bestMoveForUs, -worstNextScoreForOpp)
    }
    return bestMoveForUs, -worstNextScoreForOpp, nil
  }
}

// DFS exploration of all states at a certain depth from the given start node. 
// Once we're done, visitedStates will contain all new states we visited, and we'll return a list of leaf states. 
// These could either be terminal states or require further exploration. We should start at these states when scoring
// the play graph.
/// OK, fuck breadth first search... go back to dfs but keep the same function signature.
func exploreStates(startNode *playNode, visitedStates map[gameState]*playNode, maxDepth int) (*playNode, map[gameState]*playNode, error) {
  return exploreStatesImpl(startNode, visitedStates, make(map[gameState]*playNode, 4), 0, maxDepth)
}


func exploreStatesImpl(curNode *playNode, visitedStates map[gameState]*playNode, leaves map[gameState]*playNode, depth int, maxDepth int) (*playNode, map[gameState]*playNode, error) {
  curGs := *curNode.gs
  if visitedStates[curGs] != nil {
    // We should catch loops before we make recursive calls, so error if we detect a loop
    return nil, nil, errors.New("detected loop at beginning of recursive call: " + curNode.toString())
  }

  // Memoize the current node now so we can catch loops in recursive calls.
  visitedStates[curGs] = curNode


  if DEBUG {
    fmt.Printf("Exploring node: %s, depth: %d\n", curNode.toString(), depth)
  }

  // Sanity check: curNode should not have any children. If it does something funny is going on.
  if len(curNode.nextNodes) > 0 {
    return nil, nil, errors.New("Current node already has children, should not be explored: " + curNode.toString())
  }

  // Check for terminal states:
  if curNode.isTerminal() || curExploreNode.depth >= maxDepth {
    // This is a leaf node, add it to our output collection and continue
      if DEBUG {
        fmt.Printf(fmt.Sprintf("Found leaf node, not exploring further. cur state: %+v\n", curNode.gs))
      }
    leaves[curGs] = curNode
    return curNode, leaves, nil
  }
  // Otherwise, iterate over all possible moves
  for _, playerHand := range curNode.gs.getPlayer().getDistinctPlayableHands() {
    for _, receiverHand := range curNode.gs.getReceiver().getDistinctPlayableHands()  {
      curMove := move{playerHand, receiverHand}  

      // Make sure the gamestate gets copied....
      nextState, err := curNode.gs.copyAndPlayTurn(playerHand, receiverHand)
      if err != nil {
        return nil, nil, err
      }        
      nextNode := createPlayNodeReuseGs(nextState)

      // Here we have to check for possible loops, and if so, add the loop connection
      // in our graph and avoid recursing further.
      existingNode, exists := visitedStates[*nextNode.gs]
      if exists {
        // Sanity check
        if !existingNode.gs.equals(nextNode.gs) {
          return nil, nil, errors.New(fmt.Sprintf("Visiting states map is corrupt: visitedStates[%+v] = %s", nextNode.gs, existingNode.toString()))
        }
        // If the node already exists, we've looped around to it. Update it's pointers but don't recurse.
        if DEBUG {
          fmt.Printf(fmt.Sprintf("++ Found loop in move tree, not exploring further. cur state: %+v, loop move: %+v, next state: %+v\n", curNode.gs, curMove, existingNode.gs))
        }
        curNode.nextNodes[curMove] = existingNode
        existingNode.prevNodes[curMove] = existingNode
      } else {
        // Add the parent/child pointers and recurse on the child
        curNode.nextNodes[curMove] = nextNode
        nextNode.prevNodes[curMove] = curNode

        _, _, err := 
      }
    }
  }
  // Search is done, return the leaves we found
  return startNode, leaves, nil
}

// Problems with this:
// how do you detect and kill loops? - need to memoize game states.
// there are redundant moves: [1,1],[1,1] should have one move, not four
//
func solveDfs(curNode *playNode, visitedStates map[gameState]*playNode, leaves map[gameState]*playNode, depth int, maxDepth int) (*playNode, error) {
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
    leaves[curGs] = curNode
    if DEBUG {
      fmt.Printf("-- Returning after eliminating player1. State %+v, score: %f\n", curNode.gs, curNode.score)
    }
    return curNode, nil
  } else if (curNode.gs.player2.isEliminated()) {
    curNode.score = 1
    if curNode.gs.turn != Player2 {
      return curNode, errors.New("invalid state: player eliminated when it's not their turn")
    }
    leaves[curGs] = curNode
    if DEBUG {
      fmt.Printf("-- Returning after eliminating player2. State %+v, score: %f\n", curNode.gs, curNode.score)
    }
    return curNode, nil
  }

  if depth >= maxDepth {
    // TODO: update frontier, set heuristic score, etc.
    // Let's set a heuristic score here: +/- 0.5 for every hand you/your opponent is missing
    curNode.score = curNode.getHeuristicScore()
    leaves[curGs] = curNode
    if DEBUG {
      fmt.Printf("-- Returning after reaching maxDepth of %d. State %+v, score: %f\n", maxDepth, curNode.gs, curNode.score)
    }
    return curNode, nil 
  }

  // Recursively check all legal moves.
  // for playerHand
  player := curNode.gs.getPlayer()
  receiver := curNode.gs.getReceiver()


  for _, playerHand := range player.getDistinctPlayableHands() {
    if player.getHand(playerHand) == 0 {
      continue
    }
    for _, receiverHand := range receiver.getDistinctPlayableHands() {
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

        // Make sure we normalize the state before looking it up
        nextState.normalize()
        // HERE we have to check for possible loops, and if so, add the loop connection
        // in our graph and avoid recursing.

        existingNode, exists := visitedStates[*nextState]
        if exists {
          if DEBUG {
            fmt.Printf(fmt.Sprintf("Found loop in move tree, not recursing further. cur state: %+v, loop move: %+v, next state: %+v\n", curNode.gs, curMove, nextState))
          }
          curNode.nextNodes[curMove] = existingNode
          continue
        }

        // Add new state to cur state's children
        nextNode = createPlayNodeReuseGs(nextState)
        curNode.nextNodes[curMove] = nextNode
      }
      // Recurse, and bubble up errors
      _, err := solveDfs(nextNode, visitedStates, leaves, depth + 1, maxDepth)
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

