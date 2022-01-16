package main

import (
  "fmt"
  "log"
)

// Node for the loop metagraphs on top of the PlayNode graph
type loopNode struct {
  pn *PlayNode // The underlying PlayNode for this loopNode
  lg *loopGraph // The corresponding loop graph that this node is part of.
  nextNode *loopNode // The next node in this loop.
  prevNode *loopNode // The previous node in this loop.
}

// Loop graphs are simply a pointer to the head loop node and a size (for reference)
type loopGraph struct {
  head *loopNode
  size int
};

// Returns true if the given play node is an exit node of any loop
func isExitNode(pn *PlayNode) bool {
  return len(pn.lns) > 0
}

func PlayNodeIsInLoop(pn *PlayNode, lg *loopGraph) bool {
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

func createAndSetupLoopNode(pn *PlayNode, lg *loopGraph) *loopNode {
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
func createLoopGraphs(loops [][]*PlayNode) map[*loopGraph]int {
  loopGraphs := make(map[*loopGraph]int, len(loops))
  for loopIdx, loop := range loops {

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
    if curLoopGraph.size != len(loop) {
     log.Fatal("loop graph size does not match loop size")
   }
  }

  return loopGraphs
}

//============================================
//============ Exit Nodes ====================
//============================================

// Transforms a set of loop graphs into a map from loop graphs to their exit nodes
// TODO: need int as keys to exit nodes?
func getAllExitNodes(loopGraphs map[*loopGraph]int) map[*loopGraph]map[*PlayNode]int {
  graphsToExitNodes := make(map[*loopGraph]map[*PlayNode]int, len(loopGraphs))
  for lg, loopIdx := range loopGraphs {
    graphsToExitNodes[lg] = getExitNodes(lg, loopIdx)
  }
  return graphsToExitNodes
}

func invertExitNodesMap(loopsToExitNodes map[*loopGraph]map[*PlayNode]int) map[*PlayNode][]*loopGraph {
  exitNodesToLoopGraph := make(map[*PlayNode][]*loopGraph, len(loopsToExitNodes)) // underestimates size
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
func getExitNodes(lg *loopGraph, loopIdx int) map[*PlayNode]int {
  exitNodes := make(map[*PlayNode]int)
  for i, ln := 0, lg.head; i < lg.size; i, ln = i+1, ln.nextNode {
    pn := ln.pn
    if !PlayNodeIsInLoop(pn, lg) {
      log.Fatal("loop node loop graph does not match head graph!")
    }
    for _, nextPn := range pn.nextNodes {
      if isExitNode(nextPn) {
        // Found an exit node, add it to our set
        exitNodes[nextPn] = loopIdx
      }
    }
  }
  return exitNodes
}
// Copies the loops-to-exit-nodes map (copying map values but not pointer values)
func copyLoopsToExitNodes(loopsToExitNodes map[*loopGraph]map[*PlayNode]int) map[*loopGraph]map[*PlayNode]int {
  newLtE := make(map[*loopGraph]map[*PlayNode]int, len(loopsToExitNodes)) 
  for lg, exitNodes := range loopsToExitNodes {
    newE := make(map[*PlayNode]int, len(exitNodes))
    for e, loopIdx := range exitNodes {
      newE[e] = loopIdx
    }
    newLtE[lg] = newE
  }
  return newLtE
}

