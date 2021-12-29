package main

import (
  "fmt"
  "log"
  "os"
  "strings"
  "github.com/urfave/cli"
  "errors"
)

const DEBUG bool = true

type GameResult int8
const (
  Ongoing GameResult = iota
  Player1Wins
  Player2Wins
)

func checkGameResult(gs *gameState) GameResult {
  if gs.player1.isEliminated() {
    return Player2Wins
  } else if gs.player2.isEliminated() {
    return Player1Wins
  } else {
    return Ongoing
  }
}

func stringInputToHand(i string) (Hand, error) {
  if i == "LH" {
    return Left, nil
  } else if i == "RH" {
    return Right, nil
  } else {
    return Left, errors.New("Invalid hand " + i + ", must be either LH or RH")
  }
}


func runPlayerTurn(guiGs *gameState, curNode *playNode) (*gameState, *playNode, error) {
  fmt.Println("Your turn.")
  fmt.Println("What would you like to play?")

  // Player move
  var playerMoveStr string
  // Format: LH RH 
  fmt.Scanln(&playerMoveStr)
  // NOTE: gs might not be the same as the gs value in the curNode due to normalization!!
  playerMoveSlice := strings.Split(playerMoveStr, "->")

  playerHand, errP := stringInputToHand(playerMoveSlice[0]) 
  if errP != nil {
    return guiGs, curNode, errP
  }
  receiverHand, errR := stringInputToHand(playerMoveSlice[1]) 
  if errR != nil {
    return guiGs, curNode, errR
  }

  playerMove := move{playerHand, receiverHand}
  fmt.Println("You played: " + playerMove.toString())
  gsAfterPlayer, err := guiGs.playTurn(playerMove.playHand, playerMove.receiveHand)
  if err != nil {
    return guiGs, curNode, err
  }

  normalizedPlayerMove := normalizeMove(playerMove, curNode.gs) 
  nodeAfterPlayer, okP := curNode.nextNodes[normalizedPlayerMove]
  // NOTE: nodeAfterPlayer.gs and gsAfterPlayer may not be equal due to normalization differences, but they should
  // be equal after normalizing
  if !okP {
    return guiGs, curNode, errors.New(fmt.Sprintf("Normalized player move not found in curNode: %+v", curNode))
  }
  if DEBUG {
    normalizedGs, _, _ := gsAfterPlayer.copyAndNormalize()
    if !normalizedGs.equals(nodeAfterPlayer.gs) {
      return guiGs, curNode, errors.New(fmt.Sprintf("Normalized GUI game state and solve tree game state do not match: %+v, %+v", normalizedGs, nodeAfterPlayer.gs))
    }
    fmt.Printf("After player move: previous node: %s current node: %s, normalized move: %+v\n", curNode.toTreeString(1), nodeAfterPlayer.toTreeString(1), normalizedPlayerMove)
  }
  gsAfterPlayer.prettyPrint()
  return gsAfterPlayer, nodeAfterPlayer, nil
}

func runComputerTurn(guiGs *gameState, curNode *playNode) (*gameState, *playNode, error) {
  // Computer move
  // Need to normalize the guiGs in order to map the best move onto the current GUI
  normalizedGs, swappedPlayer1, swappedPlayer2 := guiGs.copyAndNormalize()

  normalizedComputerMove, _, errC := curNode.getBestMoveAndScore()
  if errC != nil {
    return guiGs, curNode, errC
  }
  // Denormalize the computer's move to display in the GUI
  // NOTE: ASSUMES COMPUTER IS PLAYER 2
  guiComputerMove := *normalizedComputerMove
  if swappedPlayer1 {
    guiComputerMove.receiveHand = guiComputerMove.receiveHand.invert()
  }
  if swappedPlayer2 {
    guiComputerMove.playHand = guiComputerMove.playHand.invert()
  }

  // Display the move in the GUI
  if DEBUG {
    fmt.Println("GuiGs: " + guiGs.toString())
    fmt.Println("normalizedGs: " + normalizedGs.toString())
    fmt.Println("normalizedComputerMove: " + normalizedComputerMove.toString())
    fmt.Println("guiComputerMove: " + guiComputerMove.toString())
  }
  fmt.Println("I'll play: " + guiComputerMove.toString())
  gsAfterComputer, err := guiGs.playTurn(guiComputerMove.playHand, guiComputerMove.receiveHand)
  if err != nil {
    return guiGs, curNode, err
  }

  nodeAfterComputer, okC := curNode.nextNodes[*normalizedComputerMove]
  if !okC {
    return guiGs, curNode, errors.New(fmt.Sprintf("Computer move not found in curNode: %+v", curNode))
  }
  if DEBUG {
    normalizedAfter, _, _ := gsAfterComputer.copyAndNormalize()
    if !normalizedAfter.equals(nodeAfterComputer.gs) {
      return guiGs, curNode, errors.New(fmt.Sprintf("Normalized GUI game state and solve tree game state do not match: %+v, %+v", normalizedAfter, nodeAfterComputer.gs))
    }
    fmt.Printf("After computer move: previous node: %s current node: %s, normalized move: %+v, gui move: %+v\n", curNode.toTreeString(1), nodeAfterComputer.toTreeString(1), *normalizedComputerMove, guiComputerMove)
  }
  gsAfterComputer.prettyPrint()
  return gsAfterComputer, nodeAfterComputer, nil
}

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.BoolFlag{
        Name: "dump-state",
        Usage: "language for the greeting",
      },
    },
    Action: func(c *cli.Context) error {
      gsVal := initGame()
      var gs = &gsVal
      // gsp, _  := gs.playTurn(Left, Left)
      // gs = *gsp
      var stateNode, visitedStates, solveErr = solve(gs)
      if solveErr != nil {
        fmt.Println("Error when solving: " + solveErr.Error())
        return nil
      }
      if DEBUG || c.Bool("dump-state") {
        fmt.Println(stateNode.toTreeString(9999))
      }
      fmt.Println("Let's play a game of chopsticks! You be Player 1.")
      gs.prettyPrint()
      var gameResult GameResult
      var err error = nil
      for gameResult = checkGameResult(gs); gameResult == Ongoing; gameResult = checkGameResult(gs) {
        if gs.turn == Player1 {
         gs, stateNode, err = runPlayerTurn(gs, stateNode)  
        } else {
         gs, stateNode, err = runComputerTurn(gs, stateNode)  
        }
        if err != nil {
          fmt.Println(err.Error()) 
          return nil
        }
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}