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

  if INFO {
    fmt.Printf("Created %d loops\n",len(loops))
    for lg, exitNodes := range loopGraphsToExitNodes {
      fmt.Printf("== loop graph: %p = %+v, num exit nodes: %d\n", lg, lg, len(exitNodes))
    }
  }

  // Step three: propagate scores
  // if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
  //   return nil, nil, nil, nil, err
  // }
  if INFO {
    fmt.Println(fmt.Sprintf("Root score: %f", root.score))
  }
  return root, visitedStates, leaves, loopGraphs, err
}

