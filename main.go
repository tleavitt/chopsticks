package main

import (
  "fmt"
  "log"
  "os"
  "strings"
  "github.com/urfave/cli"
  "errors"
)

const DEBUG bool = false

// func main() {
//   app := cli.NewApp()
//   app.Name = "chopsticks"
//   app.Usage = "lets play a game of chopsticks"
//   app.Action = func(c *cli.Context) error {
//     var gs = initGame()
//     // gsp, _  := gs.playTurn(Left, Left)
//     // gs = *gsp
//     solve(&gs)
//     return nil
//   }

// }


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

func runGameIteration(gs *gameState, curNode *playNode) (*gameState, *playNode, error) {
  gs.prettyPrint()
  fmt.Println("Your turn.")
  fmt.Println("What would you like to play?")

  // Player move
  var playerMoveStr string
  // Format: LH RH 
  fmt.Scanln(&playerMoveStr)
  // NOTE: gs might not be the same as the gs value in the curNode due to normalization!!
  playerMoveSlice := strings.Split(playerMoveStr, " ")
  playerHand, errP := stringInputToHand(playerMoveSlice[0]) 
  if errP != nil {
    return gs, curNode, errP
  }
  receiverHand, errR := stringInputToHand(playerMoveSlice[1]) 
  if errR != nil {
    return gs, curNode, errR
  }

  playerMove := move{playerHand, receiverHand}
  fmt.Println("You played: " + playerMove.toString())
  gsAfterPlayer, err := gs.playTurn(playerMove.playHand, playerMove.receiveHand)
  if err != nil {
    return gs, curNode, err
  }
  gsAfterPlayer.prettyPrint()


  // Computer move
  // using curNode.gs should be ok? corresponds to previous game state before player turn. enforce that?
  normalizedPlayerMove := normalizeMove(playerMove, curNode.gs) 
  nodeAfterPlayer, okP := curNode.nextNodes[normalizedPlayerMove]
  if !okP {
    return gs, curNode, errors.New(fmt.Sprintf("Normalized player move not found in curNode: %+v", curNode))
  }

  computerMove, _, errC := nodeAfterPlayer.getBestMoveAndScore()
  if errC != nil {
    return gs, curNode, errC
  }

  fmt.Println("I'll play: " + computerMove.toString())
  gsAfterComputer, err := gs.playTurn(computerMove.playHand, computerMove.receiveHand)
  nodeAfterComputer, okC :=  nodeAfterPlayer.nextNodes[*computerMove]
  if !okC {
    return gs, curNode, errors.New(fmt.Sprintf("Computer move not found in nodeAfterPlayer: %+v", nodeAfterPlayer))
  }
  return gsAfterComputer, nodeAfterComputer, err
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
      stateNode := solve(gs)
      if c.Bool("dump-state") {
        fmt.Println(stateNode.toString())
      }
      fmt.Println("Let's play a game of chopsticks! You be Player 1.")
      var gameResult GameResult
      var err error = nil
      for gameResult = checkGameResult(gs); gameResult != Ongoing; gameResult = checkGameResult(gs) {
        gs, stateNode, err = runGameIteration(gs, stateNode)
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