package main

import (
  "fmt"
  "errors"
  "math"
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
  pn.score = turnToSign(pn.gs.T) * bn.score
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

// Given a loop graph, find the current "most winning" node for each player in that loop.
func findMostWinningNodes(lg *loopGraph) (*bestNode, *bestNode, error) {
  bestPlayer1 := initBestNode()
  bestPlayer2 := initBestNode()

  for i, curNode := 0, lg.head; i < lg.size; i, curNode = i+1, curNode.nextNode {
    // Invariant: the current node should be part of the loop graph:
    if curNode.lg != lg {
      return nil, nil, errors.New(fmt.Sprintf("Node in loop graph does not point to lg: %+v, lg: %p", curNode, lg))
    }

    curScore, err := curNode.pn.computeScore(true)
    if err != nil {
      return nil, nil, err
    }

    if curNode.pn.gs.T == Player1 {
      bestPlayer1.update(curScore, curNode)
    } else {
      bestPlayer2.update(curScore, curNode)
    }
  }
  // Note: at this point we may or may not have update the best scores for each player - if there are no
  // exit nodes on a particular player's turn it will not have a best score.
  // If we haven't updated, the bestScore nodes will have a score of -2 and a bestNode of nil
  return bestPlayer1, bestPlayer2, nil
}

func enqueueLoopParent(dq *DumbQueue, ln *loopNode) {
  enqueueLoopNode(dq, ln.prevNode)
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
  nodesToScore := createDumbQueue() // Values are *loopNode

  numNodesProcessed := 0
  if b1.node != nil {
    applyScore(b1) 
    enqueueLoopParent(nodesToScore, b1.node)
    numNodesProcessed++
    if DEBUG {
      fmt.Printf("Most winning node for p1: %s (%f)\n", b1.node.pn.toString(), b1.score)
    }
  }
  if b2.node != nil {
    applyScore(b2) 
    enqueueLoopParent(nodesToScore, b2.node)
    numNodesProcessed++
    if DEBUG {
      fmt.Printf("Most winning node for p2: %s (%f)\n", b2.node.pn.toString(), b2.score)
    }
  }
  // Step three: propagate the scores up **within the loop** from the most winning nodes. 
  // Repeat BFS until all nodes in the loop are scored, OR until we run out of ways to propagate up the loop.
  // TODO: maybe a lil sketchy
  for ; nodesToScore.size > 0 && numNodesProcessed < lg.size; numNodesProcessed++ {
    curLoopNode, err := dequeueLoopNode(nodesToScore)
    if err != nil {
      return err
    }
    if DEBUG {
      fmt.Printf("Dequeued: %s\n", curLoopNode.pn.toString())
    }
    curPlayNode := curLoopNode.pn
    // If node is already scored, we've already processed it, so just skip it.
    if curPlayNode.isScored {
      if DEBUG {
        fmt.Printf("Loop node is already scored, rescoring: %s\n", curPlayNode.toString())
      }
    }
    // If all children are scored, we can score this node. If NOT all children are scored, there's another loop intersection or an unscored
    // exit node of some kind.
    if curPlayNode.allChildrenAreScored() {
      if err := curPlayNode.updateScore(); err != nil {
        return err
      }
      if DEBUG {
        fmt.Printf("Scored loop node: %+v\n", curPlayNode)
      }
    } else {
      // Killer heuristic: if the max score of a child of this node is == the score of the most winning exit, give the node that score.
      // Alternatively, if the max score of a child node is a maxed out score, give it that score as well.
      _, maxChildScoreCurPlayer, err := curPlayNode.getBestMoveAndScoreForCurrentPlayer(false, true)
      if err != nil {
        return err
      }
      var mostWinningScore float32 = 2
      if curPlayNode.gs.T == Player1 {
        if b1.node != nil {
          mostWinningScore = b1.score
        }
      } else {
        if b2.node != nil {
          mostWinningScore = b2.score
        }
      }

      // Sanity check: if this fires our most winning score code is broken, or there's some kind of loop propagation error.
      if maxChildScoreCurPlayer > mostWinningScore {
        // return fmt.Errorf("Max child score greater than most winning score: %s, %f > %f", curPlayNode.toTreeString(1), maxChildScoreCurPlayer, mostWinningScore)

        fmt.Printf("Max child score greater than most winning score: %s, %f > %f\n", curPlayNode.toTreeString(1), maxChildScoreCurPlayer, mostWinningScore)
      }

      if maxChildScoreCurPlayer >= mostWinningScore || maxChildScoreCurPlayer > 0.9 {
        curPlayNode.score = turnToSign(curPlayNode.gs.T) * maxChildScoreCurPlayer
        curPlayNode.isScored = true
        if DEBUG {
          fmt.Printf("Scored loop node based on max child score: %+v\n", curPlayNode)
        }
      }
    }

    // Always move on to the next node in the loop.
    enqueueLoopParent(nodesToScore, curLoopNode)
  }
  // Step four: go over the loop one last time and give all unscored nodes a heuristic score - they're stuck in an
  // infinite loop or are otherwise not scorable.
  applyHeuristicScores(lg)
  // At this point, all nodes in the loop should be scored (though perhaps not optimally)
  return nil
}

func applyHeuristicScores(lg *loopGraph) {
  for i, ln := 0, lg.head; i < lg.size; i, ln = i+1, ln.nextNode {
    pn := ln.pn
    // If no score, apply the heuristic score
    if !pn.isScored {
      pn.score = pn.getHeuristicScore()
      pn.isScored = true
      if DEBUG {
        fmt.Printf("Applied heuristic score to unscored loop node: %+v (%s)\n", ln, ln.pn.toString())
      }
    }
  }
}

func isScorable(node *PlayNode) bool {
  return !node.isScored && node.allChildrenAreScored()
  // return node.allChildrenAreScored()
}

func enqueueScorableParents(scorableFrontier *DumbQueue, node *PlayNode) error {
  // Assume node has been scored; 
  // enqueue all parents of node that are now scorable (i.e. they are not scored but all their children are scored)
  // Note: loop nodes are never scorable in this sense; they should never be on the scorable frontier
  for _, parentNode := range node.prevNodes {
    // Don't enqueue loop nodes, we'll handle them in the loop scoring logic.
    if isScorable(parentNode) && len(parentNode.lns) == 0 {
      enqueuePlayNode(scorableFrontier, parentNode)
    } else {
      if DEBUG {
        fmt.Println("Can't enqueue parent, not scorable: " + parentNode.toString())
      }
    }
  }
  return nil
}


func updateStateForScoredNode(curNode *PlayNode, scorableFrontier *DumbQueue, 
    remainingExitNodes map[*loopGraph]map[*PlayNode]int, exitNodesToLoopGraph map[*PlayNode][]*loopGraph) error {

  // Check if this is an exit node, and update the remainingExitNodes map if so.
  if lgs, ok := exitNodesToLoopGraph[curNode]; ok {
    for _, lg := range lgs {
      // Sanity checks 
      exitNodes, okR := remainingExitNodes[lg]
      if !okR {
        return errors.New(fmt.Sprintf("Loop graph is not present in remaining exit nodes map: %+v, %+v", lg, remainingExitNodes))
      } 
      // CurNode may have already been deleted in a previous iteration as well.
      delete(exitNodes, curNode)
    }
  }

  if err := enqueueScorableParents(scorableFrontier, curNode); err != nil {
    return err
  }
  return nil
}


// Handle enqueueing scorable parents of a loop after the loop has been scored.
func enqueueScorableParentsOfLoop(lg *loopGraph, scorableFrontier *DumbQueue, remainingExitNodes map[*loopGraph]map[*PlayNode]int, exitNodesToLoopGraph map[*PlayNode][]*loopGraph) error {
  for i, ln := 0, lg.head; i < lg.size; i, ln = i+1, ln.nextNode {
    pn := ln.pn
    // Invariant: all loop nodes should be scored when we try this.
    if !pn.isScored {
      return errors.New(fmt.Sprintf("Loop node is not scored: %+v", ln))
    }
    // Enqueue scorable parents and update our exit node graphs
    if err := updateStateForScoredNode(pn, scorableFrontier, remainingExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
  }
  return nil
}

func PlayNodeToString(pi interface{}) string {
  pn := pi.(*PlayNode)
  return pn.toString()
}

// Type safe enqueue/dequeue
func enqueuePlayNode(dq *DumbQueue, ln *PlayNode) {
  dq.enqueue(ln)
}

func dequeuePlayNode(dq *DumbQueue) (*PlayNode, error) {
  pn, err := dq.dequeue()
  if err != nil {
    return nil, err
  }
  return pn.(*PlayNode), nil
}

func scoreNodeAndUpdateState(curNode *PlayNode, scorableFrontier *DumbQueue, 
    remainingExitNodes map[*loopGraph]map[*PlayNode]int, exitNodesToLoopGraph map[*PlayNode][]*loopGraph) error {
    // Nodes on the scorable frontier must be scorable. If they're not, they might have been enqueued twice,
    // so drop them
    if !isScorable(curNode) {
      fmt.Printf("Node on frontier is not scorable: %s\n", curNode.toString())
      return nil
    }

    // Score the node
    if err := curNode.updateScore(); err != nil {
      return err
    }

    if err := updateStateForScoredNode(curNode, scorableFrontier, remainingExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
    return nil
}

// Scored frontier vs Scorable frontier:
// scored frontier consist of nodes that are likley to be able to be scored (but not a guarantee...) ???
// Scored frontier: nodes that have been scored? why do we need these? we don't
// Idea: loop over the scorable frontier until it's empty. At that point, we've scored all the nodes that we can without
// processing loops. Therefore there should be some loops that have all exit nodes scored. Score those exit nodes, then
// put the parents of the loop onto the scorable frontier, and repeat.
func propagateScores(scorableFrontier *DumbQueue, remainingExitNodes map[*loopGraph]map[*PlayNode]int, exitNodesToLoopGraph map[*PlayNode][]*loopGraph) error {
  // Drain the scorable frontier
  for loopCount := 0; scorableFrontier.size > 0; loopCount++ {
    if loopCount > 10000 {
      return errors.New("maxLoopCount exceeded, possible error in BFS graph. Frontier: %s" + scorableFrontier.toString(PlayNodeToString))
    }
    curNode, err := dequeuePlayNode(scorableFrontier)
    if err != nil {
      return err
    }
    if err := scoreNodeAndUpdateState(curNode, scorableFrontier, remainingExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
  }
  // Sanity check: we should have drained the scorable frontier now.
  if !scorableFrontier.isEmpty() {
    return errors.New(fmt.Sprintf("Scorable frontier is not empty after score propagation: %s", scorableFrontier.toString(PlayNodeToString)))
  }
  return nil
}

func createUnscoredLoopGraphMap(loopGraphs map[*loopGraph]map[*PlayNode]int) map[*loopGraph]bool {
  unscoredLoopGraphs := make(map[*loopGraph]bool, len(loopGraphs))
  for lg, _ := range loopGraphs {
    unscoredLoopGraphs[lg] = true
  }
  return unscoredLoopGraphs
}

func getLoopsWithFewestUnscoredExitNodes(unscoredLoopGraphs map[*loopGraph]bool, loopsToUnscoredExitNodes map[*loopGraph]map[*PlayNode]int) []*loopGraph {
  minUnscoredExitNodes := math.MaxInt32 
  returnLoops := []*loopGraph{}
  for lg, _ := range unscoredLoopGraphs {
    numUnscoredExitNodes := len(loopsToUnscoredExitNodes[lg])
    if numUnscoredExitNodes < minUnscoredExitNodes {
      minUnscoredExitNodes = numUnscoredExitNodes
      returnLoops = []*loopGraph{lg}
    } else if numUnscoredExitNodes == minUnscoredExitNodes {
      returnLoops = append(returnLoops, lg)
    }
  }
  return returnLoops
}

func scorePlayGraph(leaves map[*PlayNode][]*PlayNode, loopsToExitNodes map[*loopGraph]map[*PlayNode]int) error {
  // TODO also pass loops?
  // Compute the exit nodes; this map maintains all unscored exit nodes of a loop
  loopsToUnscoredExitNodes := copyLoopsToExitNodes(loopsToExitNodes)
  unscoredLoopGraphs := createUnscoredLoopGraphMap(loopsToExitNodes)
  // Need the inverse map too:
  exitNodesToLoopGraph := invertExitNodesMap(loopsToUnscoredExitNodes) 
  // Keep a running set of nodes that can (definitely?) be scored(?)
  scorableFrontier := createDumbQueue() // Values are *PlayNode

  // First, score the leaves and enqueue scorable nodes onto the scorable frontier. 
  for leaf, _ := range leaves {
    // Safety belt:
    if len(leaf.nextNodes) != 0 {
      return errors.New("Not a leaf: " + leaf.toString()) 
    }
    if err := scoreNodeAndUpdateState(leaf, scorableFrontier, loopsToUnscoredExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
  }

  if DEBUG {
    fmt.Printf("scorePlayGraph: before loop: frontier size %d, unscoredLoopGraphs size %d\n", scorableFrontier.size, len(unscoredLoopGraphs))
    fmt.Printf("Loops to exit nodes: %+v\n", loopsToUnscoredExitNodes)
    fmt.Printf("Exit nodes to loops: %+v\n", exitNodesToLoopGraph)
  }
  // Scoring iteration: consists of two steps.
  // Step 1: for all loops that have no unscored exit nodes, compute their scores and enqueue their scorable parents.
  // Step 2: propagate scores from the scorable frontier until the frontier is empty.
  // Repeat this until all nodes are scored

  // We're done if the scorable frontier is empty and all loops have been scored, so we're not done if either 
  // there are nodes on the frontier, or there are unscored loop graphs
  maxLoopCount := 2 * len(unscoredLoopGraphs)
  for loopCount := 0; !scorableFrontier.isEmpty() || len(unscoredLoopGraphs) > 0; loopCount++ {
    if loopCount > maxLoopCount {
      return errors.New("maxLoopCount exceeded in scoring iteration, frontier: " + scorableFrontier.toString(PlayNodeToString))
    }
    if DEBUG {
      fmt.Printf("scorePlayGraph: loop count %d, frontier size %d, unscoredLoopGraphs size %d\n", loopCount, scorableFrontier.size, len(unscoredLoopGraphs))
      fmt.Printf("Loops to exit nodes: %+v\n", loopsToUnscoredExitNodes)
      fmt.Printf("===== Unscored loop graphs: %+v\n", unscoredLoopGraphs)
      fmt.Printf("Exit nodes to loops: %+v\n", exitNodesToLoopGraph)
    }
    // if the frontier is empty, score a loop instead.
    if scorableFrontier.isEmpty() {

      loopGraphsToScore := getLoopsWithFewestUnscoredExitNodes(unscoredLoopGraphs, loopsToUnscoredExitNodes)
      for _, lg := range loopGraphsToScore {
        if DEBUG {
          fmt.Printf("Scoring loop graph %p\n", lg)
        }
        if err := scoreLoop(lg); err != nil {
          return err
        }
        if err := enqueueScorableParentsOfLoop(lg, scorableFrontier, loopsToUnscoredExitNodes, exitNodesToLoopGraph); err != nil {
          return err
        }
      }
      // Update our unscoredLoopGraphs map.
      for _, lg := range loopGraphsToScore {
        delete(unscoredLoopGraphs, lg)
      }
    }

    // Propagate the scores
    if err := propagateScores(scorableFrontier, loopsToUnscoredExitNodes, exitNodesToLoopGraph); err != nil {
      return err
    }
  }

  // Done!
  return nil
}

// Instead of doing fancy loop detection, just give all loop nodes a heuristic score off the bat,
// then to a score solidification down to the leaves. 
func simpleScore(root *PlayNode, loopGraphs map[*loopGraph]int, maxDepth int) error {
  for lg, _ := range loopGraphs {
    applyHeuristicScores(lg)  
  }
  solidifyScores(root, maxDepth)
  return nil
}
