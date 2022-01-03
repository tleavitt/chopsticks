package main

import (
  "fmt"
  "errors"
  "strings"
)

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

// A loop can be scored iff all of it's dependencies have been scored.

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
