package main

import (
  "log"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "chopsticks"
  app.Usage = "lets play a game of chopsticks"
  app.Action = func(c *cli.Context) error {
    var gs = initGame()
    gs.print()
    gsp, _  := gs.playTurn(Left, Left)
    gs = *gsp
    gs.print()
    return nil
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}