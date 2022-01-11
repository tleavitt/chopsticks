package main

import (
  "fmt"
  "errors"
  "log"
)

// Node for the loop metagraphs on top of the playNode graph
type loopNode struct {
  pn *playNode // The underlying playnode for this loopNode
  lg *loopGraph // The corresponding loop graph that this node is part of.
  nextNodes map[*loopNode]bool // The next node (or nodes) in this loop graph. For simple loops this has just one element.
  prevNodes map[*loopNode]bool // The previous node (or nodes) in this loop graph. For simple loops this has just one element.
}

// Loop graphs are simply a pointer to the head loop node and a size (for reference)
type loopGraph struct {
  head *loopNode
  size int
};

// Returns true if the given play node is an exit node of the given loop graph.
// Exit nodes are either not part of a loop or part of a different loop graph.
func isExitNode(pn *playNode, lg *loopGraph) bool {
  return pn.ln == nil || pn.ln.lg != lg
}

// Generics would be nice here...
func addForwardBackwardEdges(parent *loopNode, child *loopNode) {
  addForwardEdge(parent, child)
  addBackwardEdge(parent, child)
}

func addForwardEdge(parent *loopNode, child *loopNode) {
  parent.nextNodes[child] = true 
}

func addBackwardEdge(parent *loopNode, child *loopNode) {
  child.prevNodes[parent] = true
}

func createAndSetupLoopNode(pn *playNode, lg *loopGraph) *loopNode {
  ln := &loopNode{
    pn, lg, make(map[*loopNode]bool, 1), make(map[*loopNode]bool, 1),
  }
  pn.ln = ln
  lg.size++
  return ln
}

func createEmptyLoopGraph() *loopGraph {
  return &loopGraph{nil,0,}
}

func setNewLoopGraphForAll(ln *loopNode, newLg *loopGraph) {
  // Base case: we already updated this node. This means we looped back around and can return
  if ln.lg == newLg {
    return
  }
  // Update loop graph sizea
  ln.lg.size--
  ln.lg = newLg
  newLg.size++

  for nextLn, _ := range ln.nextNodes {
    setNewLoopGraphForAll(nextLn, newLg)
  } 
  // Go UP the tree as well, in case we start in the middle somewhere (should this matter at all?)
  // for prevLn, _ := range ln.prevNodes {
  //   setNewLoopGraphForAll(prevLn, newLg)
  // } 
}

// Create a set of loop graphs for the given loops.
// OK my loop code is clearly broken, giving me non-deterministic counts of exit nodes....
// or maybe it's just the number of exit nodes that's broken??
func createDistinctLoopGraphs(loops [][]*playNode) map[*loopGraph]bool {
  fmt.Printf("CreateDistinctLoopGraphs: %+v\n", loops)
  loopGraphs := make(map[*loopGraph]bool, len(loops))
  for _, loop := range loops {
    fmt.Printf("Cur loop: %+v\n", loop)
    var curLoopGraph = createEmptyLoopGraph() 
    var prevLoopNode *loopNode = nil

    loopGraphs[curLoopGraph] = true // Values are always true.

    for it, pn := range loop {
      var curLoopNode *loopNode
      if pn.ln != nil {
        // Special case: this node is already part of a loop. It could be the same loop
        // or a different one. If it's a different one we need to merge the loop graphs.
        if pn.ln.lg != curLoopGraph {
          // Remove the current loop graph and use the existing one
          delete(loopGraphs, curLoopGraph)
          fmt.Printf("Merging loop graph %p into existing loop graph %p\n", curLoopGraph, pn.ln.lg)
          // 
          // Set all the loop graph nodes in the current loop to the new loop graph
          existingLoopGraph := pn.ln.lg
          // If the current loop graph has a head defined, update it and all it's children.
        // The head could be undefined if this is the first time we're going through the update loop.
          if head := curLoopGraph.head; head != nil {
            setNewLoopGraphForAll(head, existingLoopGraph)
          }
          if curLoopGraph.size != 0 {
            fmt.Println("WARNING!!! curLoopGraph does not have zero size after updating")
          }

          // Update the curLoopGraph for future iterations
          curLoopGraph = existingLoopGraph
        }

        // Set the current loop node to the existing node.
        // Note: if this is part of the same loop graph, it means we'll end up doing
        // repeated work, but that's ok.
        curLoopNode = pn.ln
      } else {
        curLoopNode = createAndSetupLoopNode(pn, curLoopGraph)
        if curLoopGraph.head == nil {
          // First node in the loop, make it the head
          curLoopGraph.head = curLoopNode
        }
      }

      // If we have a previous node to keep track of, add the loop edges now.
      if prevLoopNode != nil { 
        addForwardBackwardEdges(prevLoopNode, curLoopNode)
      }
      // Last node, add the edges back to the loop head
      if it == len(loop) - 1 {
        addForwardBackwardEdges(curLoopNode, curLoopGraph.head)
      }
      // Iteration update
      prevLoopNode = curLoopNode
    } 
  }

  return loopGraphs
}

//============================================
//============ Exit Nodes ====================
//============================================

// Transforms a set of loop graphs into a map from loop graphs to their exit nodes
func getExitAllExitNodes(loopGraphs map[*loopGraph]bool) map[*loopGraph]map[*playNode]bool {
  graphsToExitNodes := make(map[*loopGraph]map[*playNode]bool, len(loopGraphs))
  for lg, _ := range loopGraphs {
    graphsToExitNodes[lg] = getExitNodes(lg)
  }
  return graphsToExitNodes
}

func getFirstLoopGraph(loopGraphs map[*loopGraph]bool) *loopGraph {
  for lg, _ := range loopGraphs {
    return lg
  }
  return nil
}

// Get the first mutual exit graph from the map.
func getFirstMutualExitGraphs(mutualExits map[*loopGraph][]*loopGraph) []*loopGraph {
  for _, curMutualExits := range mutualExits {
    return curMutualExits
  }
  return nil
}

func invertExitNodesMap(loopsToExitNodes map[*loopGraph]map[*playNode]bool) map[*playNode][]*loopGraph {
  exitNodesToLoopGraph := make(map[*playNode][]*loopGraph, len(loopsToExitNodes)) // underestimates size
  for lg, exitNodes := range loopsToExitNodes {
    for exitNode, _ := range exitNodes {
      var curLoops []*loopGraph
      existingLoops, ok := exitNodesToLoopGraph[exitNode]
      if ok {
        // TODO: how common is it to have multiple loop graphs for the same exit node?
        // It happens for num_fingers == 4....
        if DEBUG {
          fmt.Printf("Found exit node for %d loop graphs\n", len(existingLoops) + 1)
        }
        curLoops = append(existingLoops, lg)
      } else {
        curLoops = []*loopGraph{lg}
      }
      exitNodesToLoopGraph[exitNode] = curLoops
    }
  }
  return exitNodesToLoopGraph
}

// Get the exit nodes of the loop graph. Exit nodes are children of loop members that
// are not themselves in the same loop. (They could be in a different loop or be normal nodes.)
func getExitNodes(lg *loopGraph) map[*playNode]bool {
  exitNodes := make(map[*playNode]bool)
  getExitNodesImpl(lg.head, lg, make(map[*loopNode]bool), exitNodes)
  return exitNodes
}

func getExitNodesImpl(ln *loopNode, lg *loopGraph, visitedNodes map[*loopNode]bool, exitNodes map[*playNode]bool) {
  // Base case: we've already been here.
  if visitedNodes[ln] {
    return 
  }  
  visitedNodes[ln] = true
  if ln.lg != lg {
    log.Fatal("loop node loop graph does not match head graph!")
  }
  pn := ln.pn
  for _, nextPn := range pn.nextNodes {
    if isExitNode(nextPn, ln.lg) {
      // Found an exit node, add it to our set
      exitNodes[nextPn] = true
    }
  }

  // DFS on the loop graph
  for nextLn, _ := range ln.nextNodes {
    getExitNodesImpl(nextLn, lg, visitedNodes, exitNodes)
  }
  // Go up the graph as well?
  // DFS on the loop graph
  for prevLn, _ := range ln.prevNodes {
    getExitNodesImpl(prevLn, lg, visitedNodes, exitNodes)
  }
}

// Copies the loops-to-exit-nodes map (copying map values but not pointer values)
func copyLoopsToExitNodes(loopsToExitNodes map[*loopGraph]map[*playNode]bool) map[*loopGraph]map[*playNode]bool {
  newLtE := make(map[*loopGraph]map[*playNode]bool, len(loopsToExitNodes)) 
  for lg, exitNodes := range loopsToExitNodes {
    newE := make(map[*playNode]bool, len(exitNodes))
    for e, _ := range exitNodes {
      newE[e] = true
    }
    newLtE[lg] = newE
  }
  return newLtE
}


//===========================================================
//============ Mutual Exit Node Handlers ====================
//===========================================================

func someExitNodeIsInGraph(exitNodes map[*playNode]bool, lg *loopGraph) bool {
  for pn, _ := range exitNodes {
    if pn.ln != nil && pn.ln.lg == lg {
      return true
    }
  }
  return false
}

func findExitNodesInGraph(exitNodes map[*playNode]bool, lg *loopGraph) []*playNode {
  exitNodesInGraph := []*playNode{}
  for pn, _ := range exitNodes {
    if pn.ln != nil && pn.ln.lg == lg {
      exitNodesInGraph = append(exitNodesInGraph, pn)
    }
  }
  return exitNodesInGraph
}


func findParentsInGraph(node *playNode, lg *loopGraph) []*loopNode {
  parentsInLg := []*loopNode{}
  for _, parentNode := range node.prevNodes {
    if parentNode.ln.lg == lg {
      parentsInLg = append(parentsInLg, parentNode.ln)
    }
  }
  return parentsInLg
}

// Captures an exit node of graph1 inside of graph2 and its parent(s) in graph1
type mergeableExitNode struct {
  exitNode *loopNode
  parentNodes []*loopNode
}

// Groups all of the exitNodes1 into two groups:
// Normal exit nodes that are not members of lg2 (which are added to the normalExitNodes variable), and
// mergeable exit nodes that are mebers of lg2. For these nodes we also find the parent(s) of the exit node in lg1 so we can add new edges
//  between them.
func findMergeableExitNodes(exitNodes1  map[*playNode]bool, normalExitNodes map[*playNode]bool, lg1 *loopGraph, lg2 *loopGraph) ([]*mergeableExitNode, error) {
  exitNodes1In2 := []*mergeableExitNode{}
  for exitNode, _ := range exitNodes1 {
    if exitNode.ln != nil && exitNode.ln.lg == lg2 {
      parentsIn1 := findParentsInGraph(exitNode, lg1)
      if len(parentsIn1) == 0 {
        return nil, errors.New(fmt.Sprintf("Exit node does not have parents in graph: %+v, %p", exitNode, lg1))
      }
      exitNodes1In2 = append(exitNodes1In2, &mergeableExitNode{ exitNode.ln, parentsIn1, })
    } else {
      normalExitNodes[exitNode] = true
    }
  }
  return exitNodes1In2, nil
}


func mergeExitNodes(exitNodesToMerge []*mergeableExitNode, lg *loopGraph) error {
  if DEBUG {
    fmt.Printf("exitNodesToMerge: %+v\n", exitNodesToMerge)
  }
  for _, m := range exitNodesToMerge {
    exitNode := m.exitNode
    // Sanity check: exit node should be a part of the loop graph
    if exitNode.lg != lg {
      return errors.New(fmt.Sprintf("Exit node is not part of loop graph: %+v", exitNode))
    }
    for _, parent := range m.parentNodes {
      // Sanity check: parent node should be a part of the loop graph
      if parent.lg != lg {
        return errors.New(fmt.Sprintf("Exit node is not part of loop graph: %+v", parent))
      }

      addForwardBackwardEdges(parent, exitNode)
    }
  }
  return nil
}

func mergeGraphs(lg1 *loopGraph, lg2 *loopGraph, exitNodes1 map[*playNode]bool, exitNodes2 map[*playNode]bool) (*loopGraph, map[*playNode]bool, error) {
  // lg1 is the one that survives, for simplicity make it the bigger graph.
  if lg2.size > lg1.size {
    lg1, lg2 = lg2, lg1
  }
  // Step 1: Split the exit nodes into three groups: 
  // output exit nodes: normal exit nodes of either lg1 or lg2 
  // exitnodes from lg1 that are in lg2
  // exitnodes form lg2 that are in lg1
  normalExitNodes := make(map[*playNode]bool, len(exitNodes1) + len(exitNodes2))
  // Each element of exitNodes1
  exitNodes1In2, err := findMergeableExitNodes(exitNodes1, normalExitNodes, lg1, lg2)
  if err != nil {
    return nil, nil, err
  }

  exitNodes2In1, err := findMergeableExitNodes(exitNodes2, normalExitNodes, lg2, lg1)
  if err != nil {
    return nil, nil, err
  }

  // Step 2: merge the second loop graph into the first.
  // First set all of the lg pointers to point to lg1 instead of lg2
  setNewLoopGraphForAll(lg2.head, lg1)
  // Now, add new loop edges between all of the mergeable exit nodes and their parents.
  if err := mergeExitNodes(exitNodes1In2, lg1); err != nil {
    return nil, nil, err
  }
  if err := mergeExitNodes(exitNodes2In1, lg1); err != nil {
    return nil, nil, err
  }

  // lg1 is now the new loop graph, and normalExitNodes are now the new edges.
  return lg1, normalExitNodes, nil
}

// It could be the case that two or more loops are mutual exits of each other; i.e. an exit node of loop1 is in loop2, 
// and an exit node of loop2 is in loop1. If that's the case we have to merge the two (or more?!?) loops together
// into a single, meta-loop, otherwise we can't score them.
func mergeMutualExits(loopGraphsToExits map[*loopGraph]map[*playNode]bool) (map[*loopGraph]map[*playNode]bool, error) {
  // Make it a slice so it's easier to iterate over pairs.
  startLoops := []*loopGraph{}
  for lg, _ := range loopGraphsToExits {
    startLoops = append(startLoops, lg)
  }
  if len(startLoops) != len(loopGraphsToExits) {
    return nil, errors.New(fmt.Sprintf("startLoops invalid: %+v", startLoops))
  }

  // Maps loop graphs to lists of loop graphs that have mutual exits.
  mutualExits := make(map[*loopGraph][]*loopGraph, len(loopGraphsToExits)) 
  // For simplicity a loop graph is considered to be mutual to itself.
  for lg, _ := range loopGraphsToExits {
    mutualExits[lg] = []*loopGraph{lg}
  }

  // For each pair of graphs, check if they have mutual exits.
  for i := 0; i < len(startLoops); i++ {
    for j := i+1; j < len(startLoops); j++ {
      lg1, lg2 := startLoops[i], startLoops[j]
      exitNodes1, exitNodes2 := loopGraphsToExits[lg1], loopGraphsToExits[lg2]
      // Check if any exit nodes of 1 are in 2.
      exit1In2 := someExitNodeIsInGraph(exitNodes1, lg2)
      exit2In1 := someExitNodeIsInGraph(exitNodes2, lg1)
      // If we have BOTH, we have a pair of mutual exits
      if exit1In2 && exit2In1 {
        // Update our mutual exits list:
        existing1 := mutualExits[lg1]
        existing2 := mutualExits[lg2]
        merged := append(existing1, existing2...) // Concatenation
        // Update every entry in the merged mutual exist list
        for _, lg := range merged {
          mutualExits[lg] = merged
        }
      }
    }
  }
  if DEBUG {
    fmt.Printf("Mutual exits: %+v\n", mutualExits)
  }

  // now, merge graphs until there are no more mutual exits.
  resultLoopGraphs := make(map[*loopGraph]map[*playNode]bool, len(mutualExits))

  for curMutualExits := getFirstMutualExitGraphs(mutualExits); curMutualExits != nil; curMutualExits = getFirstMutualExitGraphs(mutualExits) {
    if len(curMutualExits) == 0 {
      return nil, errors.New("Empty mutual exit list")
    }
    // Roll up all the graphs into this first one
    lg := curMutualExits[0]
    exitNodes := loopGraphsToExits[lg]
    for i := 1; i < len(curMutualExits); i++ {
      lg2 := curMutualExits[i]
      if DEBUG {
        fmt.Printf("lg1: %p, lg2: %p\n", lg, lg2)
      }
      exitNodes2 := loopGraphsToExits[lg2]
      newLg, newExitNodes, err := mergeGraphs(lg, lg2, exitNodes, exitNodes2)
      if err != nil {
        return nil, err
      }
      lg = newLg
      exitNodes = newExitNodes
    }
    // Save our results
    resultLoopGraphs[lg] = exitNodes
    // Final step: remove all the merged graphs from our mutual exits map.
    for _, lg := range curMutualExits {
      delete(mutualExits, lg)
    }
  }

  return resultLoopGraphs, nil
}





