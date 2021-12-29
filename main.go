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

func dumpTurnInfo(gsAfterPlay *gameState, nodeAfterPlay *playNode, nodeBeforePlay *playNode, guiMove move, normalizedMove move) error {
   normalizedAfter, _, _ := gsAfterPlay.copyAndNormalize()
    if !normalizedAfter.equals(nodeAfterPlay.gs) {
      return errors.New(fmt.Sprintf("Normalized GUI game state and solve tree game state do not match: %+v, %+v", normalizedAfter, nodeAfterPlay.gs))
    }
    fmt.Println("== Begin turn info dump ==")
    fmt.Printf("previous node: %s\ncurrent node: %s\nnormalized move: %+v, gui move: %+v\n", nodeBeforePlay.toTreeString(1), nodeAfterPlay.toTreeString(1), normalizedMove, guiMove) 
    fmt.Println("== End turn info dump ==")
    return nil
}

func runPlayerTurn(gps *gamePlayState, curNode *playNode) (*gameState, *playNode, error) {
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
    return gps.state, curNode, errP
  }
  receiverHand, errR := stringInputToHand(playerMoveSlice[1]) 
  if errR != nil {
    return gps.state, curNode, errR
  }

  playerMove := move{playerHand, receiverHand}
  fmt.Println("You played: " + playerMove.toString())
  normalizedPlayerMove, err := gps.playGameTurn(playerMove)
  if err != nil {
    return 
  }
  nodeAfterPlayer, okP := curNode.nextNodes[normalizedPlayerMove]
  // NOTE: nodeAfterPlayer.gs and gsAfterPlayer may not be equal due to normalization differences, but they should
  // be equal after normalizing
  if !okP {
    return gps.state, curNode, errors.New(fmt.Sprintf("Normalized player move not found in curNode: %+v", curNode))
  }
  if DEBUG {
      if err := dumpTurnInfo(gps.state, nodeAfterPlayer, curNode, playerMove, normalizedPlayerMove); err != nil {
        return gps.state, curNode, err
      }
  }
  gps.state.prettyPrint()
  return gps.state, nodeAfterPlayer, nil
}


func runComputerTurn(guiGs *gameState, curNode *playNode) (*gameState, *playNode, error) {
  // Computer move
  // Need to normalize the guiGs in order to map the best move onto the current GUI
  normalizedGs, swappedPlayer1, swappedPlayer2 := guiGs.copyAndNormalize()

  normalizedComputerMove, _, errC := curNode.getBestMoveAndScore(true)
  if errC != nil {
    return guiGs, curNode, errC
  }
  // Denormalize the computer's move to display in the GUI
  guiComputerMove := denormalizeMove(normalizedComputerMove, swappedPlayer1, swappedPlayer2, normalizedGs.turn)

  fmt.Println("I'll play: " + guiComputerMove.toString())
  gsAfterComputer, err := guiGs.playTurn(guiComputerMove.playHand, guiComputerMove.receiveHand)
  if err != nil {
    return guiGs, curNode, err
  }

  // Invariant: normalize then play normalized then normalize should be the same as play then normalize
  if DEBUG {
    gsNormalizePlayUnnorm, err := normalizedGs.playTurn(normalizedComputerMove.playHand, normalizedComputerMove.receiveHand)
    if err != nil {
      return guiGs, curNode, err
    }
    gsNormalizePlay, _, _ := gsNormalizePlayUnnorm.copyAndNormalize()
    gsPlayNormalize, _, _ := gsAfterComputer.copyAndNormalize()
    if !gsNormalizePlay.equals(gsPlayNormalize) {
      return guiGs, curNode, errors.New(fmt.Sprintf("Invariant violation: normalization and play do not compose: normalize then play: %+v, play then normalize: %+v", gsNormalizePlay, gsPlayNormalize))
    }

  }

  nodeAfterComputer, okC := curNode.nextNodes[normalizedComputerMove]
  if !okC {
    return guiGs, curNode, errors.New(fmt.Sprintf("Computer move not found in curNode: %+v", curNode))
  }
  if DEBUG {
    if err := dumpTurnInfo(gsAfterComputer, nodeAfterComputer, curNode, guiComputerMove, normalizedComputerMove); err != nil {
      return guiGs, curNode, err
    }
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
      var stateNode, _, solveErr = solve(gs)
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