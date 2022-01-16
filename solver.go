package main

import (
  "fmt"
  "math"
  "sort"
)

const DEFAULT_MAX_DEPTH int = 150
const useSimpleScore bool = false

func getShallowestLeaf(leaves map[*PlayNode][]*PlayNode) (*PlayNode, []*PlayNode) {
  minLen := math.MaxInt32
  var minLeaf *PlayNode = nil
  var minPath []*PlayNode = nil
  for pn, path := range leaves {
    if len(path) < minLen {
      minLen = len(path)
      minLeaf = pn
      minPath = path
    }
  }
  return minLeaf, minPath
}

// Generate a play strategy given a starting game state. 
func solveRetryable(curNode *PlayNode, curPath []*PlayNode, visitedStates map[GameState]*PlayNode, maxDepth int) (*PlayNode, map[GameState]*PlayNode, map[*PlayNode][]*PlayNode, map[*loopGraph]int, error) {
  // Step one: explore all possible states, and identify loops
  root, leaves, loops, err := exploreStatesRetryable(curNode, curPath, visitedStates, maxDepth)
  if err != nil {
    return nil, nil, nil, nil, err
  }
  if INFO {
    fmt.Println(fmt.Sprintf("Generated move tree with %d nodes (%d leaves, %d loops)", len(visitedStates), len(leaves), len(loops)))

    minLeaf, minPath := getShallowestLeaf(leaves)
    fmt.Printf("Shallowest leaf: %+v (%+v)\n", minLeaf, minPath)
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
    // Step four: solidify scores until convergence
    for ;; {
      updatedScores := solidifyScores(root, maxDepth)
      if !updatedScores {
        break
      }
    }
    if INFO {
      fmt.Println(fmt.Sprintf("Root score: %f\n", root.score))
    }
  }
  return root, visitedStates, leaves, loopGraphs, err
}

type SolveCandidate struct {
  root *PlayNode
  path []*PlayNode
}

// TODO: this is probably buggy
func solveIterative(root *PlayNode, pathToRoot []*PlayNode, visitedStates map[GameState]*PlayNode, maxDepthPerIt int, iterations int) (*PlayNode, map[GameState]*PlayNode, map[*PlayNode][]*PlayNode, error) {
  solveCandidates := []*SolveCandidate{ &SolveCandidate{root, pathToRoot,} }

  for i := 0; i < iterations; i++ {
    // exit early if no more solve candidates
    if len(solveCandidates) == 0 { 
      if INFO {
        fmt.Println("Breaking early, no more solve candidates.")
      }
      break
    }
    // Sort next nodes by decreasing score for current player (i.e. best nodes for next player first)
    sort.Slice(solveCandidates, func(i, j int) bool {
      return solveCandidates[i].root.scoreForCurrentPlayer() > solveCandidates[j].root.scoreForCurrentPlayer()
    }) 
    // Iterate over explore candidates.

    nextSolveCandidates := []*SolveCandidate{}
    for _, solveCandidate := range solveCandidates {
      curRoot, curPath := solveCandidate.root, solveCandidate.path
      if curRoot.isTerminal() {
        continue
      }
      if _, contains := visitedStates[*curRoot.gs]; contains {
        fmt.Println("Leaf already visited, ignoring")
        continue
      }

      // Ignore loops? is that ok??
      _, _, leaves, _, err := solveRetryable(curRoot, curPath, visitedStates, maxDepthPerIt)
      if err != nil {
        return nil, nil, nil, err
      }
      // TODO: need to find best leaf for each player here
      // New candidates are "frontier leaves" i.e. leaves that aren't in our visited states map.
      for leaf, path := range leaves {
        nextSolveCandidates = append(nextSolveCandidates, &SolveCandidate{leaf, path,})
      }
    }
    // Propagate scores down from the root. TODO: might be costly/unnecessary to do this every time?
    if i % 2 == 1  && i != iterations - 1 {
      solidifyScores(root, math.MaxInt32)
    }

    // TODO: add alpha beta pruning here? remove the not-best nodes for each player?? 
    solveCandidates = nextSolveCandidates
  }

  solidifyScores(root, math.MaxInt32)
  // Leaves are the remaining solve candidates. TODO: this doesn't include terminal leaves, should it?
  leaves := make(map[*PlayNode][]*PlayNode, len(solveCandidates))
  for _, solveCandidate := range solveCandidates {
    leaves[solveCandidate.root] = solveCandidate.path
  }

  return root, visitedStates, leaves, nil
}

// Generate a play strategy given a starting game state. 
func solve(gs *GameState, maxDepth int) (*PlayNode, map[GameState]*PlayNode, map[*PlayNode][]*PlayNode, map[*loopGraph]int, error) {
  visitedStates := make(map[GameState]*PlayNode, 10)
  startNode := createPlayNodeCopyGs(gs)
  return solveRetryable(startNode, []*PlayNode{startNode}, visitedStates, maxDepth)
}

func solveWithIteration(gs *GameState, maxDepthPerIt int, iterations int) (*PlayNode, map[GameState]*PlayNode, map[*PlayNode][]*PlayNode, error) {
  visitedStates := make(map[GameState]*PlayNode, 10)
  startNode := createPlayNodeCopyGs(gs)
  startPath := []*PlayNode{startNode}
  return solveIterative(startNode, startPath, visitedStates, maxDepthPerIt, iterations)
}
