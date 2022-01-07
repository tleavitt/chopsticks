package main

import (
  "fmt"
  "errors"
)

type bestNode struct {
  score float32
  node *loopNode
}

func initBestNode() *bestNode {
  return &bestNode{-2, nil}  // Lower than all possible scores
}

func (bn *bestNode) update(score float32, ln *loopNode) {
  if bn.score < score {
    bn.score = score
    bn.node = ln
  }
}

func loopNodeToString(li interface{}) string {
  ln := li.(*loopNode)
  return fmt.Sprintf("%+v", ln)
}

// Type safe enqueue/dequeue
func enqueueLoopNode(dq *DumbQueue, ln *loopNode) {
  dq.enqueue(ln)
}

func dequeueLoopNode(dq *DumbQueue) (*loopNode, error) {
  ln, err := dq.dequeue()
  if err != nil {
    return nil, err
  }
  return ln.(*loopNode), nil
}

// Given a loop graph, find the "most winning" node for each player in that loop.
// All exit nodes of the graph must be scored
func findMostWinningNodes(lg *loopGraph) (*bestNode, *bestNode, error) {
  bestPlayer1 := initBestNode()
  bestPlayer2 := initBestNode()

  visitedNodes := make(map[*loopNode]bool, 4)
  // Do BFS over the loop graph
  frontier := createDumbQueue() // Values are *loopNode
  enqueueLoopNode(frontier, lg.head)

  for loopCount := 0; frontier.size > 0; loopCount++ {
    if loopCount > 10000 {
      return nil, nil, errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + frontier.toString(playNodeToString))
    }
    curNode, err := dequeueLoopNode(frontier)
    if err != nil {
      return nil, nil, err
    }

    // We've been here before, abort
    if visitedNodes[curNode] {
      continue
    }
    visitedNodes[curNode] = true

    // Invariant: the current node should be part of the loop graph:
    if curNode.lg != lg {
      return nil, nil, errors.New(fmt.Sprintf("Node in loop graph does not point to lg: %+v, lg: %p", curNode, lg))
    }

    // Score all loop exit nodes leaving from this node
    nodesToScore := make(map[move]*playNode, len(curNode.pn.nextNodes))
    for m, nextPn := range curNode.pn.nextNodes {
      // Invariant: all out edges of the nodes should either be part of the same loop
      // or not be part of a loop at all.
      if nextPn.ln != nil {
        if nextPn.ln.lg != lg {
          return nil, nil, errors.New(fmt.Sprintf("Next play node in loop graph is part of a different loop graph: %+v, current lg: %p, new lg %p", nextPn, lg, nextPn.ln.lg))
        }
      } else {
        // Else this is an exit node, evaluate its score
        nodesToScore[m] = nextPn
      }
    }

    // If there are exit nodes: update the best move score. Otherwise just keep chugging along, we don't want to update
    // the scores to zeros or some other value
    if len(nodesToScore) > 0 {
      _, curScore, err := getBestMoveAndScore(nodesToScore, false, false)
      if err != nil {
        return nil, nil, err
      }

      if curNode.pn.gs.turn == Player1 {
        bestPlayer1.update(curScore, curNode)
      } else {
        bestPlayer2.update(curScore, curNode)
      }
    }

    // Move on to the next nodes in the loop graph
    for nextNode, _ := range curNode.nextNodes {
      enqueueLoopNode(frontier, nextNode)
    }
  }
  // Note: at this point we may or may not have update the best scores for each player - if there are no
  // exit nodes on a particular player's turn it will not have a best score.
  // If we haven't updated, the bestScore nodes will have a score of -2 and a bestNode of nil
  return bestPlayer1, bestPlayer2, nil
}

func scoreLoop(lg *loopGraph) {
  // Invariant one: all edges of the loop should either be part of the same loop
  // or not be part of a loop at all. 
  // Invariant two: all exit nodes of the loop must be scored.
  // Step one: find the most "winning" exit edges of the loop **for each player**.
  //  -- winning means: best score for current player. Most winning states are +1/Player1Turn, -1/Player2Turn. If both
  //     exist we have to score both
  // Step two: score the most winning node(s) of the loop. The score is simply equal to the most winning edge.
  // Step three: propagate the scores up **within the loop** from the most winning nodes. Repeat BFS until all nodes in the loop are scored.

}



func scoreAndEnqueueParents(node *playNode, frontier *DumbQueue) error {
  if err := node.updateScore(false); err != nil {
    return err
  }
  if DEBUG {
    fmt.Println("Computed score for node: " + node.toString())
  }
  for _, parentNode := range node.prevNodes {
    // Safety belt: only enqueue nodes if they haven't been scored already
    if !parentNode.isScored {
      frontier.enqueue(parentNode)
    }
  }
  return nil
}

// A loop can be scored iff all of it's dependencies have been scored.

func playNodeToString(pi interface{}) string {
  pn := pi.(*playNode)
  return pn.toString()
}

// Here we have to use BFS 
// Goal after this function returns: all nodes that are scoreable have scores.
// NOTE: if the leaves map is incomplete this function will go into an infinite loop
// Todo: loop detection? 
func propagateScores(leaves map[gameState]*playNode, maxLoopCount int) error {
  fmt.Println("Started propagateScores")
  // Queue of states to explore
  frontier := createDumbQueue() // Values are *playNode
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

  // while
  for loopCount := 0; frontier.size > 0; loopCount++ {
    if loopCount > maxLoopCount {
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + frontier.toString(playNodeToString))
    }
    curNodeI, err := frontier.dequeue()
    if err != nil {
      return err
    }
    curNode := curNodeI.(*playNode)
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
      frontier.enqueue(curNode)
    }
  }
  // At this point all nodes should be scored
  return nil
}
