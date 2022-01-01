package main

import (
  "fmt"
  "errors"
  "strings"
)

const DEFAULT_MAX_DEPTH int = 15
// 
func solve(gs *gameState) (*playNode, map[gameState]*playNode, map[gameState]*playNode, error) {
  visitedStates := make(map[gameState]*playNode, 10)
  // Step one: explore all possible states
  root, leaves, err := exploreStates(createPlayNodeCopyGs(gs), visitedStates, DEFAULT_MAX_DEPTH)
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

// DFS exploration of all states at a certain depth from the given start node. 
// Once we're done, visitedStates will contain all new states we visited, and we'll return a list of leaf states. 
// These could either be terminal states or require further exploration. We should start at these states when scoring
// the play graph.
/// OK, fuck breadth first search... go back to dfs but keep the same function signature.
func exploreStates(startNode *playNode, visitedStates map[gameState]*playNode, maxDepth int) (*playNode, map[gameState]*playNode, error) {
  return exploreStatesImpl(startNode, visitedStates, make(map[gameState]*playNode, 4), 0, maxDepth)
}

func wireUpParentChildPointers(parent *playNode, child *playNode, m move) {
  parent.nextNodes[m] = child 
  child.prevNodes[m] = parent
}

// Breadth First Search is trickier here because you can't guarantee that starting an iteration with a node you've visited before is an error.
// So you have to handle that case specially.
func exploreStatesImpl(curNode *playNode, curPath map[gameState]bool, visitedStates map[gameState]*playNode, leaves map[gameState]*playNode, depth int, maxDepth int) (*playNode, map[gameState]*playNode, error) {
  curGs := *curNode.gs
  if visitedStates[curGs] != nil {
    // We should always catch loops before we make recursive calls, so error if we detect a loop
    return nil, nil, errors.New("detected loop at beginning of recursive call: " + curNode.toString())
  }

  // Memoize the current node now so we can catch loops in recursive calls.
  visitedStates[curGs] = curNode

  if DEBUG {
    fmt.Printf("Exploring node: %s, depth: %d\n", curNode.toString(), depth)
  }

  // Sanity check: curNode should not have any children. If it does something funny is going on.
  if len(curNode.nextNodes) > 0 {
    return nil, nil, errors.New("Current node already has children, should not be explored: " + curNode.toString())
  }

  // Check for terminal states:
  if curNode.isTerminal() || depth >= maxDepth {
    // This is a leaf node, add it to our output collection and continue
      if DEBUG {
        fmt.Printf(fmt.Sprintf("Found leaf node, not exploring further. cur state: %+v\n", curNode.gs))
      }
    leaves[curGs] = curNode
    return curNode, leaves, nil
  }
  // Otherwise, iterate over all possible moves
  for _, playerHand := range curNode.gs.getPlayer().getDistinctPlayableHands() {
    for _, receiverHand := range curNode.gs.getReceiver().getDistinctPlayableHands()  {
      curMove := move{playerHand, receiverHand}  

      // Make sure the gamestate gets copied....
      nextState, err := curNode.gs.copyAndPlayTurn(playerHand, receiverHand)
      if err != nil {
        return nil, nil, err
      }        
      nextNode := createPlayNodeReuseGs(nextState)

      // Here we have to check for possible intersections and loops. 
      // An intersection is when the current path leads to a state we've already explored somewhere else in our search.
      // A loop is an intersection where the existing state is on our current path.
      // If we find an intersection, we need to add parent/child pointers from the curNode to the existing node to complete
      // the graph
      // In addition, if we find a loop, we need to mark the current node as a "leaf" so we can score it directly later. Otherwise
      // the scoring step will never finish because the parents of the loop will never have all of their children scored.
      // TODO: more correct from a scoring perspective would be to collapse all loops into single "supernodes" from a scoring perspective. 
      // (although turns become tricky there...)
      // All nodes in a supernode would get the same score (?)
      // Note that loops necessarily must have the same number of hands throughout so they will have the same heuristic score.
      existingNode, exists := visitedStates[*nextNode.gs]
      if exists {
        // Sanity check
        if !existingNode.gs.equals(nextNode.gs) {
          return nil, nil, errors.New(fmt.Sprintf("Visiting states map is corrupt: visitedStates[%+v] = %s", nextNode.gs, existingNode.toString()))
        }
        if DEBUG {
          fmt.Printf(fmt.Sprintf("++ Found loop in move tree, marking current node as leaf and not exploring further. cur state: %+v, loop move: %+v, next state: %+v\n", curNode.gs, curMove, existingNode.gs))
        }
        leaves[curGs] = curNode
        wireUpParentChildPointers(curNode, existingNode, curMove)
      } else {
        // Add the parent/child pointers and recurse on the child
        wireUpParentChildPointers(curNode, nextNode, curMove)
        _, _, err := exploreStatesImpl(nextNode, visitedStates, leaves, depth + 1, maxDepth)
        if err != nil {
          return nil, nil, err
        }
      }
    }
  }
  // Search is done, return the leaves we found
  return curNode, leaves, nil
}

func chanToString(ch <-chan *playNode) string {
  var sb strings.Builder
  sb.WriteString("chan{\n")
  for node := range ch {
    sb.WriteString(node.toString())
    sb.WriteString("\n")
  }
  sb.WriteString("}")
  return sb.String()
}


func scoreAndEnqueueParents(node *playNode, frontier chan<- *playNode) error {
  if err := node.updateScore(); err != nil {
    return err
  }
  if DEBUG {
    fmt.Println("Computed score for node: " + node.toString())
  }
  for _, parentNode := range node.prevNodes {
    // Safety belt: only enqueue nodes if they haven't been scored already
    if !parentNode.isScored {
      frontier <- parentNode
    }
  }
  return nil
}

// Here we have to use BFS 
// Goal after this function returns: all nodes that are scoreable have scores.
// NOTE: if the leaves map is incomplete this function will go into an infinite loop
// Todo: loop detection? 
func propagateScores(leaves map[gameState]*playNode, maxLoopCount int, frontierSize int) error {
  fmt.Println("Started propagateScores")
  // Queue of states to explore
  frontier := make(chan *playNode, frontierSize) // Maximum int32, needed otherwise pushing a value will block....
  // Score the leaves and add their immediate parents to the frontier
  for _, leaf := range leaves {
    // Safety belt:
    if len(leaf.nextNodes) != 0 {
      return errors.New("Not a leaf: " + leaf.toString()) 
    }
    if err := scoreAndEnqueueParents(leaf, frontier); err != nil {
      return err
    }
  }

  for loopCount, frontierHasValues := 0, true; frontierHasValues; loopCount++ {
    if loopCount > maxLoopCount {
      close(frontier)
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + chanToString(frontier))
    }
    // Ugh it's kind of ugly to use channels as queues here
    select {
    case curNode, ok := <-frontier:
      if ok {
        // If this node is already scored, skip this step. It's parents have already been enqueued in a previous iteration.
        if curNode.isScored {
          continue
        }
        // A node can be scored iff all it's children have been scored.
        if curNode.allChildrenAreScored() {
          if err := scoreAndEnqueueParents(curNode, frontier); err != nil {
            return err
          }
        } else {
          // Put it back on the pile, if the leaves are complete we'll get to all it's children eventually
          // Important: frontier NEEDS to be FIFO otherwise we won't make progress.
          frontier <- curNode
        }
      } else {
        return errors.New("Frontier channel closed!")
      }
    default:
      if DEBUG {
        fmt.Println("Exhausted states to explore")
      }
      frontierHasValues = false
    }
  }
  // At this point all nodes should be scored
  return nil
}
