package main

import (
  "fmt"
)

const DEFAULT_MAX_DEPTH int = 25

// Generate a play strategy given a starting game state. 
func solve(gs *gameState, maxDepth int) (*playNode, map[gameState]*playNode, map[gameState]*playNode, map[*loopGraph]bool, error) {
  // Step one: explore all possible states, and identify loops
  visitedStates := make(map[gameState]*playNode, 10)
  root, leaves, loops, err := exploreStates(createPlayNodeCopyGs(gs), visitedStates, DEFAULT_MAX_DEPTH)
  if err != nil {
    return nil, nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves, %d loops)", len(visitedStates), len(leaves), len(loops)))
  }

  // Step two: build loop graphs
  loopGraphs := createDistinctLoopGraphs(loops)
  if INFO {
    fmt.Println(fmt.Sprintf("Created %d loop graphs from %d loops", len(loopGraphs), len(loops)))
  }

  // Step three: propagate scores
  if err := scorePlayGraph(leaves, loopGraphs); err != nil {
    return nil, nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, loopGraphs, err
}

