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
// TODO: could also implement by looping over the exit nodes instead?
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
      if isExitNode(nextPn, lg) {
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
    applyScore(b1) 
    enqueueLoopParents(nodesToScore, b1.node)
    if DEBUG {
      fmt.Printf("Most winning node for p1: %s (%f)\n", b1.node.pn.toString(), b1.score)
    }
  }
  if b2.node != nil {
    applyScore(b2) 
    enqueueLoopParents(nodesToScore, b2.node)
    if DEBUG {
      fmt.Printf("Most winning node for p2: %s (%f)\n", b2.node.pn.toString(), b2.score)
    }
  }
  // Step three: propagate the scores up **within the loop** from the most winning nodes. 
  // Repeat BFS until all nodes in the loop are scored, OR until we run out of ways to propagate up the loop.
  for loopCount := 0; nodesToScore.size > 0; loopCount++ {
    if loopCount > 10000 {
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. nodesToScore: %s" + nodesToScore.toString(loopNodeToString))
    }
    curLoopNode, err := dequeueLoopNode(nodesToScore)
    fmt.Printf("Dequeued: %s\n", curLoopNode.pn.toString())
    if err != nil {
      return err
    }
    curPlayNode := curLoopNode.pn
    // If node is already scored, we've already processed it, so skip.
    if curPlayNode.isScored {
      if DEBUG {
        fmt.Printf("Loop node is already scored, skipping: %s\n", curPlayNode.toString())
      }
      continue
    }
    // If all children are scored, we can score this node. If NOT all children are scored, there's another branch of the loop we have to climb up
    // OR this is an infinite loop, so don't enqueue parents.
    // EDIT: this is not necessarily true for loop junctions... we could get to a point where the other branch of the loop junction is isolated
    //  so we'll never make progress.
    // So we use another trick: if any child of a node has a score of +1 (from the perspective of the node) then that node must have a score of +1,
    // since no other child can have a higher score. 
    // There are probably still dragons lurking in here... but hopefully they are rare ¯\_(ツ)_/¯
    if curPlayNode.allChildrenAreScored() {
      if err := curPlayNode.updateScore(); err != nil {
        return err
      }
      if DEBUG {
        fmt.Printf("Scored loop node: %+v\n", curPlayNode)
      }
      enqueueLoopParents(nodesToScore, curLoopNode)
    } else if maxChildScoreCurPlayer := curPlayNode.maxChildScoreForPlayer(); maxChildScoreCurPlayer > 0.9 {
      curPlayNode.score = turnToSign(curPlayNode.gs.turn) * maxChildScoreCurPlayer
      curPlayNode.isScored = true
      if DEBUG {
        fmt.Printf("Scored loop node based on max child score: %+v\n", curPlayNode)
      }
      enqueueLoopParents(nodesToScore, curLoopNode)
    }

  }
  // Step four: go over the loop one last time and give all unscored nodes a heuristic score - they're stuck in an
  // infinite loop or are otherwised borked somehow.
  scoreInfiniteLoops(lg)
  // At this point, all nodes in the loop should be scored (though perhaps not optimally)
  return nil
}

func scoreInfiniteLoops(lg *loopGraph) {
  scoreInfiniteLoopsImpl(lg.head, make(map[*loopNode]bool))
}

func scoreInfiniteLoopsImpl(ln *loopNode, visitedNodes map[*loopNode]bool) {
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
      fmt.Printf("Applied heuristic score to unscored loop node: %+v (%s)\n", ln, ln.pn.toString())
    }
  }

  // Continue DFS
  for childNode, _ := range ln.nextNodes {
    scoreInfiniteLoopsImpl(childNode, visitedNodes)
  }
}

func isScorable(node *playNode) bool {
  return !node.isScored && node.allChildrenAreScored()
}

func enqueueScorableParents(scorableFrontier *DumbQueue, node *playNode) {
  // Assume node has been scored; 
  // enqueue all parents of node that are now scorable (i.e. they are not scored but all their children are scored)
  for _, parentNode := range node.prevNodes {
    if isScorable(parentNode) {
      enqueuePlayNode(scorableFrontier, parentNode)
    }
  }
}

func enqueueAllScorableParents(lg *loopGraph, scorableFrontier *DumbQueue) error {
  return enqueueAllScorableParentsImpl(lg.head, scorableFrontier, make(map[*loopNode]bool))
}

func enqueueAllScorableParentsImpl(ln *loopNode, scorableFrontier *DumbQueue, visitedNodes map[*loopNode]bool) error {
  // Base case: already been here
  if visitedNodes[ln] {
    return nil 
  }
  visitedNodes[ln] = true
  pn := ln.pn
  // Invariant: all loop nodes should be scored.
  if !pn.isScored {
    return errors.New(fmt.Sprintf("Loop node is not scored: %+v", ln))
  }
  // Enqueue scorable parents
  enqueueScorableParents(scorableFrontier, pn)
  // Recurse
  for childNode, _ := range ln.nextNodes {
    enqueueAllScorableParentsImpl(childNode, scorableFrontier, visitedNodes)
  }
  return nil
}

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

func scoreNodeAndUpdateState(curNode *playNode, scorableFrontier *DumbQueue, 
    remainingExitNodes map[*loopGraph]map[*playNode]bool, exitNodesToLoopGraph map[*playNode][]*loopGraph) error {
  // Nodes on the scorable frontier must be scorable.
  if !isScorable(curNode) {
    return errors.New(fmt.Sprintf("Node on frontier is not scorable: %s", curNode.toString()))
  }

  // Score the node
  if err := curNode.updateScore(); err != nil {
    return err
  }

  // Check if this is an exit node, and update the remainingExitNodes map if so.
  if lgs, ok := exitNodesToLoopGraph[curNode]; ok {
    for _, lg := range lgs {
      // Sanity checks 
      exitNodes, okR := remainingExitNodes[lg]
      if !okR {
        return errors.New(fmt.Sprintf("Loop graph is not present in remaining exit nodes map: %+v, %+v", lg, remainingExitNodes))
      } 
      if _, okE := exitNodes[curNode]; !okE {
        return errors.New(fmt.Sprintf("Exit nodes does not contain play node: %+v, %+v", curNode, exitNodes))
      }

      delete(exitNodes, curNode)
    }
  }

  enqueueScorableParents(scorableFrontier, curNode) 
  return nil
}

// Scored frontier vs Scorable frontier:
// scored frontier consist of nodes that are likley to be able to be scored (but not a guarantee...) ???
// Scored frontier: nodes that have been scored? why do we need these? we don't
// Idea: loop over the scorable frontier until it's empty. At that point, we've scored all the nodes that we can without
// processing loops. Therefore there should be some loops that have all exit nodes scored. Score those exit nodes, then
// put the parents of the loop onto the scorable frontier, and repeat.
func propagateScores(scorableFrontier *DumbQueue, remainingExitNodes map[*loopGraph]map[*playNode]bool, exitNodesToLoopGraph map[*playNode][]*loopGraph) error {
  // Drain the scorable frontier
  for loopCount := 0; scorableFrontier.size > 0; loopCount++ {
    if loopCount > 10000 {
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + scorableFrontier.toString(playNodeToString))
    }
    curNode, err := dequeuePlayNode(scorableFrontier)
    if err != nil {
      return err
    }
    scoreNodeAndUpdateState(curNode, scorableFrontier, remainingExitNodes, exitNodesToLoopGraph)
  }
  // Sanity check: we should have drained the scorable frontier now.
  if !scorableFrontier.isEmpty() {
    return errors.New(fmt.Sprintf("Scorable frontier is not empty after score propagation: %s", scorableFrontier.toString(playNodeToString)))
  }
  return nil
}

func createUnscoredLoopGraphMap(loopGraphs map[*loopGraph]bool) map[*loopGraph]bool {
  unscoredLoopGraphs := make(map[*loopGraph]bool, len(loopGraphs))
  for lg, _ := range loopGraphs {
    unscoredLoopGraphs[lg] = true
  }
  return unscoredLoopGraphs
}

func scorePlayGraph(leaves map[gameState]*playNode, loopGraphs map[*loopGraph]bool) error {
  // Compute the exit nodes; this map maintains all unscored exit nodes of a loop
  loopsToExitNodes := getExitAllExitNodes(loopGraphs)
  unscoredLoopGraphs := createUnscoredLoopGraphMap(loopGraphs)
  // Need the inverse map too:
  exitNodesToLoopGraph := invertExitNodesMap(loopsToExitNodes) 
  // Keep a running set of nodes that can (definitely?) be scored(?)
  scorableFrontier := createDumbQueue() // Values are *playNode

  // First, score the leaves and enqueue scorable nodes onto the scorable frontier. 
  for _, leaf := range leaves {
    // Safety belt:
    if len(leaf.nextNodes) != 0 {
      return errors.New("Not a leaf: " + leaf.toString()) 
    }
    scoreNodeAndUpdateState(leaf, scorableFrontier, loopsToExitNodes, exitNodesToLoopGraph)
  }

  if DEBUG {
    fmt.Printf("scorePlayGraph: before loop: frontier size %d, unscoredLoopGraphs size %d\n", scorableFrontier.size, len(unscoredLoopGraphs))
    fmt.Printf("Loops to exit nodes: %+v\n", loopsToExitNodes)
    fmt.Printf("Exit nodes to loops: %+v\n", exitNodesToLoopGraph)
  }
  // Scoring iteration: consists of two steps.
  // Step 1: for all loops that have no unscored exit nodes, compute their scores and enqueue their scorable parents.
  // Step 2: propagate scores from the scorable frontier until the frontier is empty.
  // Repeat this until all nodes are scored

  // We're done if the scorable frontier is empty and all loops have been scored, so we're not done if either 
  // there are nodes on the frontier, or there are unscored loop graphs
  for loopCount := 0; !scorableFrontier.isEmpty() || len(unscoredLoopGraphs) > 0; loopCount++ {
    if loopCount > 2 {
      return errors.New("maxLoopCount exceeded in scoring iteration, frontier: %s" + scorableFrontier.toString(playNodeToString))
    }
    if DEBUG {
      fmt.Printf("scorePlayGraph: loop count %d, frontier size %d, unscoredLoopGraphs size %d\n", loopCount, scorableFrontier.size, len(unscoredLoopGraphs))
      fmt.Printf("Loops to exit nodes: %+v\n", loopsToExitNodes)
      fmt.Printf("Exit nodes to loops: %+v\n", exitNodesToLoopGraph)
    }
    // Find all loops with no unscored exit nodes, and score them.
    curScoredLoopGraphs := []*loopGraph{}
    for lg, _ := range unscoredLoopGraphs {
      exitNodes := loopsToExitNodes[lg]
      if len(exitNodes) == 0 {
        if err := scoreLoop(lg); err != nil {
          return err
        }
        // now: lg.isScored == true
        if err := enqueueAllScorableParents(lg, scorableFrontier); err != nil {
          return err
        }
        // Record that we scored this graph
        curScoredLoopGraphs = append(curScoredLoopGraphs, lg)
      }
    }

    // Update our unscoredLoopGraphs map.
    for _, lg := range curScoredLoopGraphs {
      delete(unscoredLoopGraphs, lg)
    }

    // Propagate the scores
    if err := propagateScores(scorableFrontier, loopsToExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
  }

  // Done!
  return nil
}
