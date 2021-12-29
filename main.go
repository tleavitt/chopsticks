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
   normalizedAfter := gsAfterPlay.copyAndNormalize()
    if !normalizedAfter.equals(nodeAfterPlay.gs) {
      return errors.New(fmt.Sprintf("Normalized GUI game state and solve tree game state do not match: %+v, %+v", normalizedAfter, nodeAfterPlay.gs))
    }
    fmt.Println("== Begin turn info dump ==")
    fmt.Printf("previous node: %s\ncurrent node: %s\nnormalized move: %+v, gui move: %+v\n", nodeBeforePlay.toTreeString(1), nodeAfterPlay.toTreeString(1), normalizedMove, guiMove) 
    fmt.Println("== End turn info dump ==")
    return nil
}

func validateGpsAndNode(gps *gamePlayState, curNode *playNode) error {
  if !gps.normalizedState.equals(curNode.gs) {
    return errors.New(fmt.Sprintf("GPS normalized state and solve node state do not match: %+v, %+v", gps.normalizedState, curNode.gs))
  } else {
    return nil
  }
}

func runPlayerTurn(gps *gamePlayState, curNode *playNode) (*playNode, error) {
  fmt.Println("Your turn.")
  fmt.Println("What would you like to play?")

  // Player move
  var playerMoveStr string
  // Format: LH RH 
  fmt.Scanln(&playerMoveStr)
  // NOTE: gs might not be the same as the gs value in the curNode due to normalization!!
  playerMoveSlice := strings.Split(playerMoveStr, "->")

  playerHand, err := stringInputToHand(playerMoveSlice[0]) 
  if err != nil {
    return curNode, err
  }
  receiverHand, err := stringInputToHand(playerMoveSlice[1]) 
  if err != nil {
    return curNode, err
  }

  playerMove := move{playerHand, receiverHand}
  fmt.Println("You played: " + playerMove.toString())
  normalizedPlayerMove, err := gps.playGameTurn(playerMove)
  if err != nil {
    return curNode, err
  }
  nodeAfterPlayer, okP := curNode.nextNodes[normalizedPlayerMove]
  // NOTE: nodeAfterPlayer.gs and gsAfterPlayer may not be equal due to normalization differences, but they should
  // be equal after normalizing
  if !okP {
    return curNode, errors.New(fmt.Sprintf("Normalized player move not found in curNode: %+v", curNode))
  }
  if DEBUG {
      if err := dumpTurnInfo(gps.state, nodeAfterPlayer, curNode, playerMove, normalizedPlayerMove); err != nil {
        return curNode, err
      }
  }
  gps.state.prettyPrint()
  return nodeAfterPlayer, nil
}


func runComputerTurn(gps *gamePlayState, curNode *playNode) (*playNode, error) {
  // Computer move
  // Need to normalize the guiGs in order to map the best move onto the current GUI
  normalizedComputerMove, _, err := curNode.getBestMoveAndScore(true)
  if err != nil {
    return curNode, err
  }

  guiComputerMove, err := gps.playNormalizedTurn(normalizedComputerMove)
  if err != nil {
    return curNode, err
  }

  fmt.Println("I'll play: " + guiComputerMove.toString())

  // // Invariant: normalize then play normalized then normalize should be the same as play then normalize
  // if DEBUG {
  //   gsNormalizePlayUnnorm, err := normalizedGs.playTurn(normalizedComputerMove.playHand, normalizedComputerMove.receiveHand)
  //   if err != nil {
  //     return curNode, err
  //   }
  //   gsNormalizePlay, _, _ := gsNormalizePlayUnnorm.copyAndNormalize()
  //   gsPlayNormalize, _, _ := gsAfterComputer.copyAndNormalize()
  //   if !gsNormalizePlay.equals(gsPlayNormalize) {
  //     return guiGs, curNode, errors.New(fmt.Sprintf("Invariant violation: normalization and play do not compose: normalize then play: %+v, play then normalize: %+v", gsNormalizePlay, gsPlayNormalize))
  //   }
  // }

  nodeAfterComputer, okC := curNode.nextNodes[normalizedComputerMove]
  if !okC {
    return curNode, errors.New(fmt.Sprintf("Computer move not found in curNode: %+v", curNode))
  }
  if DEBUG {
    if err := dumpTurnInfo(gps.state, nodeAfterComputer, curNode, guiComputerMove, normalizedComputerMove); err != nil {
      return curNode, err
    }
  }
  gps.state.prettyPrint()
  return nodeAfterComputer, nil
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
      gs := &gsVal

      var stateNode, _, _, solveErr = solve(gs)
      if solveErr != nil {
        fmt.Println("Error when solving: " + solveErr.Error())
        return nil
      }
      if DEBUG || c.Bool("dump-state") {
        fmt.Println(stateNode.toTreeString(9999))
      }
      fmt.Println("Let's play a game of chopsticks! You be Player 1.")
      gs.prettyPrint()

      gps := createGamePlayState(gs)
      var gameResult GameResult
      var err error = nil
      for gameResult = checkGameResult(gps.state); gameResult == Ongoing; gameResult = checkGameResult(gps.state) {
        if DEBUG {
          if err := validateGpsAndNode(gps, stateNode); err != nil {
            return err
          }
        }
        if gps.state.turn == Player1 {
          stateNode, err = runPlayerTurn(gps, stateNode)  
        } else {
          stateNode, err = runComputerTurn(gps, stateNode)  
        }
        if err != nil {
          return err
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