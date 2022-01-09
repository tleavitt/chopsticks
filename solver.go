package main

import (
  "fmt"
)

const DEFAULT_MAX_DEPTH int = 15
// 
func solve(gs *gameState) (*playNode, map[gameState]*playNode, map[gameState]*playNode, map[*loopGraph]bool, error) {
  visitedStates := make(map[gameState]*playNode, 10)
  // Step one: explore all possible states, and identify loops
  root, leaves, loops, err := exploreStates(createPlayNodeCopyGs(gs), visitedStates, DEFAULT_MAX_DEPTH)
  if err != nil {
    return nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves, %d loops)", len(visitedStates), len(leaves), len(loops)))
  }

  // Step two: build loop graphs
  loopGraphs := createDistinctLoopGraphs(loops)

  // Step three: propagate scores
  if err := scorePlayGraph(leaves, loopGraphs); err != nil {
    return nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, loopGraphs, err
}

