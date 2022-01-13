package main

import (
  "fmt"
)

const DEFAULT_MAX_DEPTH int = 25
const useSimpleScore bool = false

// Generate a play strategy given a starting game state. 
func solveRetryable(curNode *playNode, curPath []*playNode, visitedStates map[gameState]*playNode, maxDepth int) (*playNode, map[gameState]*playNode, map[*playNode][]*playNode, map[*loopGraph]int, error) {
  // Step one: explore all possible states, and identify loops
  root, leaves, loops, err := exploreStatesRetryable(curNode, curPath, visitedStates, DEFAULT_MAX_DEPTH)
  if err != nil {
    return nil, nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves, %d loops)", len(visitedStates), len(leaves), len(loops)))
  }

  // Step two: build loop graphs and find exit nodes
  loopGraphs := createLoopGraphs(loops)
  if useSimpleScore {
    simpleScore(root, loopGraphs, maxDepth)
  } else {
    loopGraphsToExitNodes := getAllExitNodes(loopGraphs)

    if INFO {
      for lg, exitNodes := range loopGraphsToExitNodes {
        fmt.Printf("== loop graph: %p = %+v, num exit nodes: %d\n", lg, lg, len(exitNodes))
      }
    }

    // Step three: propagate scores
    if err := scorePlayGraph(leaves, loopGraphsToExitNodes); err != nil {
      return nil, nil, nil, nil, err
    }
    // Step four: do one more pass down the tree and touch up any inaccuracies...
    solidifyScores(root, maxDepth)
    if INFO {
      fmt.Println(fmt.Sprintf("Root score: %f", root.score))
    }
  }
  return root, visitedStates, leaves, loopGraphs, err
}

// Generate a play strategy given a starting game state. 
func solve(gs *gameState, maxDepth int) (*playNode, map[gameState]*playNode, map[*playNode][]*playNode, map[*loopGraph]int, error) {
  visitedStates := make(map[gameState]*playNode, 10)
  startNode := createPlayNodeCopyGs(gs)
  return solveRetryable(startNode, []*playNode{startNode}, visitedStates, maxDepth)
}