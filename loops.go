package main

import ("fmt")

// Node for the loop metagraphs on top of the playNode graph
type loopNode struct {
  pn *playNode // The underlying playnode for this loopNode
  lg *loopGraph // The corresponding loop graph that this node is part of.
  nextNodes map[*loopNode]bool // The next node (or nodes) in this loop graph. For simple loops this has just one element.
  prevNodes map[*loopNode]bool // The previous node (or nodes) in this loop graph. For simple loops this has just one element.
}

// Loop graphs are simply a pointer to the head loop node.
type loopGraph struct {
  head *loopNode
};

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

func createDistinctLoopGraphs(loops [][]*playNode) map[*loopGraph]bool {
  fmt.Printf("CreateDistinctLoopGraphs: %+v\n", loops)
  loopGraphs := make(map[*loopGraph]bool, len(loops))
  for _, loop := range loops {
    fmt.Printf("Cur loop: %+v\n", loop)
    var curLoopGraph = createEmptyLoopGraph() 
    var prevLoopNode *loopNode = nil
    loopGraphs[curLoopGraph] = true

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






