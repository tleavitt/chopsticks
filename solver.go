package main

import (
  "fmt"
  "errors"
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

  // Step two: build loop graphs, find exit nodes, and merge when encessary
  loopGraphs := createDistinctLoopGraphs(loops)
  loopGraphsToExitNodes := getExitAllExitNodes(loopGraphs)
  initialNumLg := len(loopGraphsToExitNodes)
  // Need to iteratively merge loops until no more loops can be merged, since the loop edges can change after each merge.
  for it := 0; ;it++ {
    // Safety belt
    if it > 100 {
      return nil, nil, nil, nil, errors.New("too many merge loop iterations")
    }
    prevNumLoop := len(loopGraphsToExitNodes)
    var err error = nil;
    if loopGraphsToExitNodes, err = mergeMutualExits(loopGraphsToExitNodes); err != nil {
      return nil, nil, nil, nil, err
    }
    if prevNumLoop == len(loopGraphsToExitNodes) {
      break
    }
  }

  if INFO {
    fmt.Println(fmt.Sprintf("Created %d consolidated loop graphs from %d unmerged loop graphs (%d loops)", len(loopGraphsToExitNodes), initialNumLg, len(loops)))
  }

  // Step three: propagate scores
  if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
    return nil, nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, loopGraphs, err
}

