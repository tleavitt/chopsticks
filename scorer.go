package main

import (
  "fmt"
  "errors"
)



// func findMostWinningNodes(lg *loopGraph) (*loopNode, *loopNode, error) {
//   var curNode *loopNode = lg.head
//   var bestPlayer1Node *loopNode = nil
//   var bestPlayer1Score float32 = -2 // Lower than all possible scores
//   var bestPlayer2Node *loopNode = nil
//   var bestPlayer2Score float32 = 2 // Higher than all possible scores

//   visitedNodes := make(map[*loopNode]bool, 4)

// }

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
