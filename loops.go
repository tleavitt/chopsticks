package main

import ("fmt")

// Node for the loop metagraphs on top of the playNode graph
type loopNode struct {
  pn *playNode // The underlying playnode for this loopNode
  lg *loopGraph // The corresponding loop graph that this node is part of.
  nextNodes map[*loopNode]bool // The next node (or nodes) in this loop graph. For simple loops this has just one element.
  prevNodes map[*loopNode]bool // The previous node (or nodes) in this loop graph. For simple loops this has just one element.
}

// Loop graphs are simply a pointer to the head loop node
type loopGraph struct {
  head *loopNode
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
  return ln
}

func createEmptyLoopGraph() *loopGraph {
  return &loopGraph{nil,}
}

func setNewLoopGraphForAll(ln *loopNode, newLg *loopGraph) {
  // Base case: we already updated this node. This means we looped back around and can return
  if ln.lg == newLg {
    return
  }
  ln.lg = newLg
  for nextLn, _ := range ln.nextNodes {
    setNewLoopGraphForAll(nextLn, newLg)
  } 
}

// Create a set of loop graphs for the given loops.
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
        // Special case: this node is already part of another loop
        // Remove the current loop graph and use the existing one
        delete(loopGraphs, curLoopGraph)
        // Set all the loop graph nodes in the current loop to the new loop graph
        existingLoopGraph := pn.ln.lg
        // If the current loop graph has a head defined, update it and all it's children.
      // The head could be undefined if this is the first time we're going through the update loop.
        if head := curLoopGraph.head; head != nil {
          setNewLoopGraphForAll(head, existingLoopGraph)
        }

        // Update the curLoopGraph for future iterations
        curLoopGraph = existingLoopGraph
        // Set the current loop node to the existing node
        curLoopNode = pn.ln
      } else {
        curLoopNode = createAndSetupLoopNode(pn, curLoopGraph)
      }

      if it == 0 {
        // First node in the loop, make it the head
        curLoopGraph.head = curLoopNode
      } else {
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

func getFirstLoopGraph(loopGraphs map[*loopGraph]bool) *loopGraph {
  for lg, _ := range loopGraphs {
    return lg
  }
  return nil
}

// Transforms a set of loop graphs into a map from loop graphs to their exit nodes
func getExitAllExitNodes(loopGraphs map[*loopGraph]bool) map[*loopGraph]map[*playNode]bool {
  graphsToExitNodes := make(map[*loopGraph]map[*playNode]bool, len(loopGraphs))
  for lg, _ := range loopGraphs {
    graphsToExitNodes[lg] = getExitNodes(lg)
  }
  return graphsToExitNodes
}

func invertExitNodesMap(loopsToExitNodes map[*loopGraph]map[*playNode]bool) map[*playNode][]*loopGraph {
  exitNodesToLoopGraph := make(map[*playNode][]*loopGraph, len(loopsToExitNodes)) // underestimates size
  for lg, exitNodes := range loopsToExitNodes {
    for exitNode, _ := range exitNodes {
      var curLoops []*loopGraph
      existingLoops, ok := exitNodesToLoopGraph[exitNode]
      if ok {
        // TODO: how common is it to have multiple loop graphs for the same exit node?
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
  getExitNodesImpl(lg.head, make(map[*loopNode]bool), exitNodes)
  return exitNodes
}

func getExitNodesImpl(ln *loopNode, visitedNodes map[*loopNode]bool, exitNodes map[*playNode]bool) {
  // Base case: we've already been here.
  if visitedNodes[ln] {
    return 
  }  
  visitedNodes[ln] = true
  pn := ln.pn
  for _, nextPn := range pn.nextNodes {
    if isExitNode(nextPn, ln.lg) {
      // Found an exit node, add it to our set
      exitNodes[nextPn] = true
    }
  }

  // DFS on the loop graph
  for nextLn, _ := range ln.nextNodes {
    getExitNodesImpl(nextLn, visitedNodes, exitNodes)
  }
}






