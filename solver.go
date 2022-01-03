package main

import (
  "fmt"
  "errors"
)

const DEFAULT_MAX_DEPTH int = 15
// 
func solve(gs *gameState) (*playNode, map[gameState]*playNode, map[gameState]*playNode, error) {
  visitedStates := make(map[gameState]*playNode, 10)
  // Step one: explore all possible states
  root, leaves, _, err := exploreStates(createPlayNodeCopyGs(gs), visitedStates, DEFAULT_MAX_DEPTH)
  if err != nil {
    return nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves)", len(visitedStates), len(leaves)))
  }
  // Step two: propagate scores
  if err := propagateScores(leaves, 5 * len(visitedStates), 10 * len(visitedStates)); err != nil {
    return nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, err
}

// Note: the node must not be a leaf (i.e. it must have children) or this function will fail
func (node *playNode) getBestMoveAndScore(log bool) (move, float32, error) {
  // Our best move is the move that puts our opponent in the worst position.
  // The score of the current node is the negative of the score of our opponent in the node after our best move.
  var worstNextScoreForOpp float32 = 2 // This is an impossible score, so we should always trigger an update in the loop.
  var bestMoveForUs move // This should always get updated.
  if log {
    fmt.Printf("-- Running getBestMoveAndScore() for %+v\n", node.gs)
  }
  for nextMove, nextNode := range node.nextNodes {
    if !nextNode.isScored {
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
    return bestMoveForUs, 0, errors.New(fmt.Sprintf("getBestMoveAndScore: play node is invalid: %+s; worst next score for opp: %f", node.toString(), worstNextScoreForOpp))
  } else {
    // Note the negative sign!! worst score for opp is the best score for us.
    if log {
      fmt.Printf("-- result %+v, %f\n", bestMoveForUs, -worstNextScoreForOpp)
    }
    return bestMoveForUs, -worstNextScoreForOpp, nil
  }
}
