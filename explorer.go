package main

import (
  "fmt"
  "errors"
  "sort"
)

// DFS exploration of all states at a certain depth from the given start node. 
// Once we're done, visitedStates will contain all new states we visited, and we'll return a list of leaf states. 
// These could either be terminal states or require further exploration. We should start at these states when scoring
// the play graph.
/// OK, fuck breadth first search... go back to dfs but keep the same function signature.
func exploreStates(startNode *PlayNode, visitedStates map[gameState]*PlayNode, maxDepth int) (*PlayNode, map[*PlayNode][]*PlayNode, [][]*PlayNode, error) {
  return exploreStatesImpl(startNode, []*PlayNode{startNode}, visitedStates, make(map[*PlayNode][]*PlayNode, 4), make([][]*PlayNode, 0, 4), maxDepth, 0)
}

func exploreStatesRetryable(startNode *PlayNode, curPath []*PlayNode, visitedStates map[gameState]*PlayNode, maxDepth int) (*PlayNode, map[*PlayNode][]*PlayNode, [][]*PlayNode, error) {
  if INFO {
    fmt.Printf("Exploring state %s\n", startNode.toString())
  }
  return exploreStatesImpl(startNode, curPath, visitedStates, make(map[*PlayNode][]*PlayNode, 4), make([][]*PlayNode, 0, 4), maxDepth, len(curPath) - 1)
}

// Yes, O(N) search. Whatever, it's probably fine
func findNodeInPath(node *PlayNode, path []*PlayNode) int {
  for i, curNode := range path {
    if node == curNode {
      return i
    }
  }
  return -1
}

// Egad, go slices are fucking awful
func copyPath(path []*PlayNode) []*PlayNode {
  newPath := make([]*PlayNode, len(path)) 
  copy(newPath, path)
  return newPath
}

type exploreCandidate struct {
  pn *PlayNode
  moveToPn *move
  heuristic float32
}

func exploreStatesImpl(curNode *PlayNode, curPath []*PlayNode, visitedStates map[gameState]*PlayNode, leaves map[*PlayNode][]*PlayNode, loops [][]*PlayNode, maxDepth int, baseDepth int) (*PlayNode, map[*PlayNode][]*PlayNode, [][]*PlayNode, error) {
  // Sanity check: curNode should be the last node of the path
  if curPath[len(curPath) - 1] != curNode {
    return nil, nil, nil, errors.New(fmt.Sprintf("current path is invalid, last node should be %+v: %+v", curNode, curPath))
  }
  depth := len(curPath) - baseDepth

  curGs := *curNode.gs
  if visitedStates[curGs] != nil {
    // We should always catch intersections before we make recursive calls, so error if we detect an intersection
    return nil, nil, nil, errors.New("detected intersection at beginning of recursive call: " + curNode.toString())
  }

  // Memoize the current node now so we can catch intersections in recursive calls.
  visitedStates[curGs] = curNode

  if DEBUG {
    fmt.Printf("Exploring node: %s, depth: %d\n", curNode.toString(), depth)
  }

  // Check if we've hit the max depth - if so mark this node as a "frontier node", i.e. a non-terminal leaf, and DON'T save it 
  // to our visited states map (since we haven't visited it.)
  // TODO: too support retryable-ness this should happen earlier.
  if depth >= maxDepth {
    // This is a leaf node, add it to our output collection and continue
    if INFO {
      fmt.Printf(fmt.Sprintf("Hit max depth, not exploring further. cur state: %+v, depth %d\n", curNode.gs, depth))
    }
    // Mark this as a leaf, but _NOT_ an explored state
    leaves[curNode] = copyPath(curPath)
    return curNode, leaves, loops, nil
  }


  // Sanity check: curNode should not have any children. If it does something funny is going on.
  if len(curNode.nextNodes) > 0 {
    return nil, nil, nil, errors.New("Current node already has children, should not be explored: " + curNode.toString())
  }


  // Check for terminal states, i.e. TERMINAL leaf nodes. We also need to save these because we need them for scoring.
  if curNode.isTerminal() {
    // This is a leaf node, add it to our output collection and continue
      if INFO {
        fmt.Printf(fmt.Sprintf("Found leaf node, not exploring further. cur state: %+v, depth %d\n", curNode.gs, depth))
      }
    leaves[curNode] = copyPath(curPath)
    return curNode, leaves, loops, nil
  }
  // Otherwise, iterate over all possible moves

  // Micro-opt: recurse on the best nodes for the next player first, according to their heuristic
  exploreCandidates := []*exploreCandidate{}
  for _, playerHand := range curNode.gs.getPlayer().getDistinctPlayableHands() {
    for _, receiverHand := range curNode.gs.getReceiver().getDistinctPlayableHands()  {
      curMove := move{playerHand, receiverHand}  

      // Make sure the gamestate gets copied....
      nextState, err := curNode.gs.copyAndPlayTurn(playerHand, receiverHand)
      if err != nil {
        return nil, nil, nil, err
      }        
      nextNode := createPlayNodeReuseGs(nextState)

      exploreCandidates = append(exploreCandidates, &exploreCandidate{
        nextNode, &curMove, nextNode.getHeuristicScoreForCurrentPlayer(),
      })
    }
  }
  // Sort next nodes by decreasing heuristic (i.e. best nodes for next player first)
  sort.Slice(exploreCandidates, func(i, j int) bool {
    return exploreCandidates[i].heuristic > exploreCandidates[j].heuristic
  })

  // Recurse
  for _, toExplore := range exploreCandidates {
    nextNode := toExplore.pn
    curMove := toExplore.moveToPn
    // Here we have to check for possible intersections and loops. 
    // An intersection is when the current path leads to a state we've already explored somewhere else in our search.
    // A loop is an intersection where the existing state is on our current path.
    // If we find an intersection, we need to add parent/child pointers from the curNode to the existing node to complete
    // the graph
    // In addition, if we find a loop, we need to store the loop in our "loops" return value.
    existingNode, exists := visitedStates[*nextNode.gs]
    if exists {
      // Sanity check
      if !existingNode.gs.equals(nextNode.gs) {
        return nil, nil, nil, errors.New(fmt.Sprintf("Visiting states map is corrupt: visitedStates[%+v] = %s", nextNode.gs, existingNode.toString()))
      }
      addParentChildEdges(curNode, existingNode, *curMove)
      if DEBUG {
        fmt.Printf(fmt.Sprintf("++ Found intersection in move tree, not exploring further. cur node: %s, loop move: %+v, next node: %s\n", curNode.toString(), curMove, existingNode.toString()))
      }
      // Check for loops
      if loopIdx := findNodeInPath(existingNode, curPath); loopIdx != -1 {
        curLoop := copyPath(curPath[loopIdx:])
        if DEBUG {
          fmt.Printf("++++ Found LOOP in move tree, saving loop for later: %+v\n", curLoop)
        }
        loops = append(loops, curLoop)
      }
    } else {
      // Add the parent/child pointers and recurse on the child
      addParentChildEdges(curNode, nextNode, *curMove)
      // append the latest node to our current path
      // oldLen := len(curPath)
      // exploreCandidates = append(exploreCandidates, &exploreCandidate{
      //   nextNode, nextNode.getHeuristicScoreForCurrentPlayer(),
      // })
      nextPath := append(curPath, nextNode)
      _, _, newLoops, err := exploreStatesImpl(nextNode, nextPath, visitedStates, leaves, loops, maxDepth, baseDepth)
      loops = newLoops
      if err != nil {
        return nil, nil, nil, err
      }
      // Remove the latest node from our path to keep recursing (not necessary?)
      // curPath = curPath[:oldLen]
    }
  }

  // Search is done, return the leaves we found
  return curNode, leaves, loops, nil
}

// Explore the game tree and correct any incorrect scores.
func solidifyScores(startNode *PlayNode, maxDepth int) bool {
  return solidifyScoresImpl(startNode, make(map[*PlayNode]bool, 4), 0, maxDepth)
}

func solidifyScoresImpl(curNode *PlayNode, visitedNodes map[*PlayNode]bool, depth int, maxDepth int) bool {
  // Abort after we hit the maximum depth.
  if depth >= maxDepth { 
    return false
  }
  // Base case: we've been here before, return. Means we're in a loop or an intersection.
  if visitedNodes[curNode] {
    return false
  }
  visitedNodes[curNode] = true

  // Allow unscored nodes as well
  // if !curNode.isScored {
  //   return fmt.Errorf("Unscored node when solidifying scores: %s", curNode.toString())
  // }

  // First, solidify scores for all children. For leaves this will be empty.
  someChildUpdatedScore := false
  for _, childNode := range curNode.nextNodes {
    if solidifyScoresImpl(childNode, visitedNodes, depth-1, maxDepth) {
      someChildUpdatedScore = true
    }
  }

  // Then, update the score for the current node.
  prevScore := curNode.score
  curNode.updateScore()
  if curNode.score != prevScore {
    if DEBUG {
      fmt.Printf("Updated score for node %s, previous score: %f\n", curNode.toString(), prevScore)
    }
    return true
  } else {
    return someChildUpdatedScore
  }
}