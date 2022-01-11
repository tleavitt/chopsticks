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
  nextNode *loopNode // The next node in this loop.
  prevNode *loopNode // The previous node in this loop.
}

// Loop graphs are simply a pointer to the head loop node and a size (for reference)
type loopGraph struct {
  head *loopNode
  size int
};

// Returns true if the given play node is an exit node of the given loop graph.
// Exit nodes are nodes whose parent is in a graph, but they themselves are not in a graph.
func isExitNode(pn *playNode) bool {
  return len(pn.lns) == 0
}

func playNodeIsInLoop(pn *playNode, lg *loopGraph) bool {
  for _, ln := range pn.lns {
    if ln.lg == lg {
      return true
    }
  }
  return false
}

// Generics would be nice here...
func addForwardBackwardEdges(parent *loopNode, child *loopNode) {
  addForwardEdge(parent, child)
  addBackwardEdge(parent, child)
}

func addForwardEdge(parent *loopNode, child *loopNode) {
  parent.nextNode = child
}

func addBackwardEdge(parent *loopNode, child *loopNode) {
  child.prevNode = parent
}

func createAndSetupLoopNode(pn *playNode, lg *loopGraph) *loopNode {
  ln := &loopNode{
    pn, lg, nil, nil,
  }
  pn.lns = append(pn.lns, ln)
  lg.size++
  return ln
}

func createEmptyLoopGraph() *loopGraph {
  return &loopGraph{nil,0,}
}

// Create a set of loop graphs for the given loops; map values are the corresponding index into the loops slice.
func createLoopGraphs(loops [][]*playNode) map[*loopGraph]int {
  fmt.Printf("createLoopGraphs: %+v\n", loops)
  loopGraphs := make(map[*loopGraph]int, len(loops))
  for loopIdx, loop := range loops {
    fmt.Printf("Cur loop: %+v\n", loop)

    var curLoopGraph = createEmptyLoopGraph() 
    var prevLoopNode *loopNode = nil

    loopGraphs[curLoopGraph] = loopIdx 

    for it, pn := range loop {
      var curLoopNode *loopNode
      curLoopNode = createAndSetupLoopNode(pn, curLoopGraph)

      if curLoopGraph.head == nil {
        // First node in the loop, make it the head
        curLoopGraph.head = curLoopNode
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

// Get the exit nodes of the loop. Exit nodes are children of loop members that
// are not themselves in the same loop. (They could be in a different loop or be normal nodes.)
func getExitNodes(lg *loopGraph, loop []*playNode) map[*playNode]bool {
  exitNodes := make(map[*playNode]bool)
  for _, pn := range loop {
    if !playNodeIsInLoop(pn, lg) {
      log.Fatal("loop node loop graph does not match head graph!")
    }
    for _, nextPn := range pn.nextNodes {
      if isExitNode(nextPn) {
        // Found an exit node, add it to our set
        exitNodes[nextPn] = true
      }
    }
  }
  return exitNodes
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

