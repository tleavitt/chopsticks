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

func applyScore(bn *bestNode) {
  pn := bn.node.pn
  pn.score = turnToSign(pn.gs.turn) * bn.score
  pn.isScored = true
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
      return nil, nil, errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + frontier.toString(loopNodeToString))
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
      // or scored. Note: the next edge could be  part of a different loop; if that's the case it must be scored.
      if nextPn.ln == nil || nextPn.ln.lg != lg {
        // If we're here, next node is either not in a loop or in a different loop. In both cases it is an exit node 
        // and must be scored.
        if !nextPn.isScored {
          return nil, nil, errors.New(fmt.Sprintf("Exit node of loop is not scored: %+v, current ln: %+v", nextPn, curNode))
        }
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

func enqueueLoopParents(dq *DumbQueue, ln *loopNode) {
  for parentNode, _ := range ln.prevNodes {
    enqueueLoopNode(dq, parentNode)
  }
}

func scoreLoop(lg *loopGraph) error {
  // Step one: find the most "winning" exit edges of the loop **for each player**.
  //  -- winning means: best score for current player. Most winning states are +1/Player1Turn, -1/Player2Turn. If both
  //     exist we have to score both
  b1, b2, err := findMostWinningNodes(lg)
  if err != nil {
    return err
  }
  // Step two: score the most winning node(s) of the loop. The score is simply equal to the most winning edge.
  // Add the loop-parents of the most winning nodes to 
  nodesToScore := createDumbQueue() // Values are *loopNode
  if b1.node != nil {
    if DEBUG {
      fmt.Printf("Most winnign node for p1: %+v (%f)", b1.node, b1.score)
    }
    applyScore(b1) 
    enqueueLoopParents(nodesToScore, b1.node)
  }
  if b2.node != nil {
    if DEBUG {
      fmt.Printf("Most winnign node for p2: %+v (%f)", b2.node, b2.score)
    }
    applyScore(b2) 
    enqueueLoopParents(nodesToScore, b2.node)
  }
  // Step three: propagate the scores up **within the loop** from the most winning nodes. 
  // Repeat BFS until all nodes in the loop are scored, OR until we run out of ways to propagate up the loop.
  for loopCount := 0; nodesToScore.size > 0; loopCount++ {
    if loopCount > 10000 {
      return nil, nil, errors.New("maxLoopCount exceeded, possible error in BFS graph. nodesToScore: %s" + nodesToScore.toString(loopNodeToString))
    }
    curLoopNode, err := dequeueLoopNode(nodesToScore)
    if err != nil {
      return err
    }
    curPlayNode := curLoopNode.pn
    // If node is already scored, we've already processed it, so skip.
    if curPlayNode.isScored {
      if DEBUG {
        fmt.Printf("Loop node is already scored, skipping: %s", curPlayNode.toString())
      }
      continue
    }
    // If all children are scored, we can score this node. If NOT all children are scored, there's another branch of the loop we have to climb up
    // OR this is an infinite loop, so don't enqueue parents.
    if curPlayNode.allChildrenAreScored() {
      if err := curPlayNode.updateScore(false); err != nil {
        return err
      }
      if DEBUG {
        fmt.Printf("Scored loop node: %s", curPlayNode.toString())
      }
      enqueueLoopParents(nodesToScore, curLoopNode)
    }
  }
  // Step four: go over the loop one last time and give all unscored nodes a heuristic score - they're stuck in an
  // infinite loop
  scoreInfiniteLoops(ln)
  // At this point, all nodes in the loop should be scored. Return
}

func scoreInfiniteLoops(lg *loopGraph) {
  scoreInfiniteLoopsImpl(lg.head, make(map[*loopNode]bool))
}

func scoreInfiniteLoopsImpl(ln *loopNode, visiteNodes map[*loopNode]bool) {
  // Base case: already been here
  if visitedNodes[ln] {
    return
  }
  visitedNodes[ln] = true
  pn := ln.pn
  // If no score, this is an infinite loop, so apply the heuristic score
  if !pn.isScored {
    pn.score = pn.getHeuristicScore()
    pn.isScored = true
    if DEBUG {
      fmt.Printf("Applied heuristic score to unscored loop node: %+v (%s)", ln, ln.pn.toString())
    }
  }

  // Continue DFS
  for childNode, _ := range ln.nextNodes {
    scoreInfiniteLoopsImpl(childNOde, visitedNodes)
  }
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
      enqueuePlayNode(frontier, parentNode)
    }
  }
  return nil
}

// A loop can be scored iff all of it's dependencies have been scored.

func playNodeToString(pi interface{}) string {
  pn := pi.(*playNode)
  return pn.toString()
}

// Type safe enqueue/dequeue
func enqueuePlayNode(dq *DumbQueue, ln *playNode) {
  dq.enqueue(ln)
}

func dequeuePlayNode(dq *DumbQueue) (*playNode, error) {
  pn, err := dq.dequeue()
  if err != nil {
    return nil, err
  }
  return pn.(*playNode), nil
}

// Here we have to use BFS 
// Goal after this function returns: all nodes that are scoreable have scores.
// NOTE: if the leaves map is incomplete this function will go into an infinite loop
// Todo: loop detection? 
func propagateScores(leaves map[gameState]*playNode) error {
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
    if loopCount > 10000 {
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + frontier.toString(playNodeToString))
    }
    curNode, err := dequeuePlayNode(frontier)
    if err != nil {
      return err
    }
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
      enqueuePlayNode(frontier, curNode)
    }
  }
  // At this point all nodes should be scored
  return nil
}
