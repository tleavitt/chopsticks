package main

import (
  "fmt"
  "log"
  "os"
  "github.com/urfave/cli"
  "time"
)

// const DEBUG bool = true
var DEBUG bool = false
const INFO bool = true

func main() {
  app := &cli.App{
    Flags: []cli.Flag {
      &cli.BoolFlag{
        Name: "dump-state",
        Usage: "language for the greeting",
      },
    },
    Commands: []*cli.Command{
      {
        Name:    "cli",
        Aliases: []string{"c"},
        Usage:   "play chopsticks on the cli",
        Action: func(c *cli.Context) error {
          gs := initGame()
          start := time.Now()
          var stateNode, _, _, _, solveErr = solve(gs, DEFAULT_MAX_DEPTH)
          duration := time.Since(start)
          fmt.Println("Computed solve state in:") // 10s of ms, hot damn golang is fast
          fmt.Println(duration)
          if solveErr != nil {
            fmt.Println("Error when solving: " + solveErr.Error())
            return nil
          }
          if c.Bool("dump-state") {
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
            // if false {
              stateNode, err = runPlayerTurn(gps, stateNode)  
            } else {
              time.Sleep(1 * time.Second)
              stateNode, err = runComputerTurn(gps, stateNode)  
            }
            if err != nil {
              return err
            }
            fmt.Printf("Cur state: %s\n", stateNode.toString())
          }

          // Game over!
          fmt.Println("Game over!")
          if gameResult == Player1Wins {
            fmt.Println("You win!")
          } else {
            fmt.Println("I win!")
          }
          return nil
        },
      },
      {
        Name:    "serve",
        Aliases: []string{"s"},
        Usage:   "play chopsticks with a browser",
        Action:  func(c *cli.Context) error {
          return nil
        },
      },

  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}