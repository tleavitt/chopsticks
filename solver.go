package main

import (
  "fmt"
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
  if err := propagateScores(leaves, 5 * len(visitedStates)); err != nil {
    return nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, err
}

