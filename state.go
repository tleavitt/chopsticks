package main

import (
  "fmt"
)

	
type player struct {
    lh int
    rh  int
}

type gameState struct {
	player1 player
	player2 player
}

func (gs gameState) print() {
	fmt.Println("==================================")
	fmt.Println("==           Player 2           ==")
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", gs.player2.lh, gs.player2.rh)
	fmt.Println("==------------------------------==")
	fmt.Println("==           Player 1           ==")
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", gs.player1.lh, gs.player1.rh)
	fmt.Println("==================================")
}

func initGame() gameState {
	return gameState{
		player{1, 1},	
		player{1, 1},	
	}	
}