package main

import (
  "fmt"
  "errors"
)

// DFS exploration of all states at a certain depth from the given start node. 
// Once we're done, visitedStates will contain all new states we visited, and we'll return a list of leaf states. 
// These could either be terminal states or require further exploration. We should start at these states when scoring
// the play graph.
/// OK, fuck breadth first search... go back to dfs but keep the same function signature.
func exploreStates(startNode *playNode, visitedStates map[gameState]*playNode, maxDepth int) (*playNode, map[gameState]*playNode, [][]*playNode, error) {
  return exploreStatesImpl(startNode, []*playNode{startNode}, visitedStates, make(map[gameState]*playNode, 4), make([][]*playNode, 0), 0, maxDepth)
}

// Yes, O(N) search. Whatever, it's probably fine
func findNodeInPath(node *playNode, path []*playNode) int {
  for i, curNode := range path {
    if node == curNode {
      return i
    }
  }
  return -1
}

// Egad, go slices are fucking awful
func copyPath(path []*playNode) []*playNode {
  newPath := make([]*playNode, len(path)) 
  copy(newPath, path)
  return newPath
}


func exploreStatesImpl(curNode *playNode, curPath []*playNode, visitedStates map[gameState]*playNode, leaves map[gameState]*playNode, loops [][]*playNode, depth int, maxDepth int) (*playNode, map[gameState]*playNode, [][]*playNode, error) {
  // Sanity check: curNode should be the last node of the path
  if curPath[len(curPath) - 1] != curNode {
    return nil, nil, nil, errors.New(fmt.Sprintf("current path is invalid, last node should be %+v: %+v", curNode, curPath))
  }
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

  // Sanity check: curNode should not have any children. If it does something funny is going on.
  if len(curNode.nextNodes) > 0 {
    return nil, nil, nil, errors.New("Current node already has children, should not be explored: " + curNode.toString())
  }

  // Check for terminal states:
  if curNode.isTerminal() || depth >= maxDepth {
    // This is a leaf node, add it to our output collection and continue
      if DEBUG {
        fmt.Printf(fmt.Sprintf("Found leaf node, not exploring further. cur state: %+v\n", curNode.gs))
      }
    leaves[curGs] = curNode
    return curNode, leaves, loops, nil
  }
  // Otherwise, iterate over all possible moves
  for _, playerHand := range curNode.gs.getPlayer().getDistinctPlayableHands() {
    for _, receiverHand := range curNode.gs.getReceiver().getDistinctPlayableHands()  {
      curMove := move{playerHand, receiverHand}  

      // Make sure the gamestate gets copied....
      nextState, err := curNode.gs.copyAndPlayTurn(playerHand, receiverHand)
      if err != nil {
        return nil, nil, nil, err
      }        
      nextNode := createPlayNodeReuseGs(nextState)

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
        if DEBUG {
          fmt.Printf(fmt.Sprintf("++ Found intersection in move tree, marking current node as leaf and not exploring further. cur state: %+v, loop move: %+v, next state: %+v\n", curNode.gs, curMove, existingNode.gs))
        }
        addParentChildEdges(curNode, existingNode, curMove)
        // Check for loops
        if loopIdx := findNodeInPath(existingNode, curPath); loopIdx != -1 {
          curLoop := copyPath(curPath[loopIdx:])
          if DEBUG {
            fmt.Printf("++++ Found LOOP in move tree, saving loop for later: %+v\n", curLoop)
            fmt.Printf("Current loops: %+v\n", loops)
          }
          loops = append(loops, curLoop)
          if DEBUG {
            fmt.Printf("New loops: %+v\n", loops)
          }
        }
      } else {
        // Add the parent/child pointers and recurse on the child
        addParentChildEdges(curNode, nextNode, curMove)
        // append the latest node to our current path
        oldLen := len(curPath)
        nextPath := append(curPath, nextNode)
        _, _, newLoops, err := exploreStatesImpl(nextNode, nextPath, visitedStates, leaves, loops, depth + 1, maxDepth)
        loops = newLoops
        if err != nil {
          return nil, nil, nil, err
        }
        // Remove the latest node from our path to keep recursing
        curPath = curPath[:oldLen]
      }
    }
  }
  // Search is done, return the leaves we found
  return curNode, leaves, loops, nil
}
