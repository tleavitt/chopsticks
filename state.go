package main

import (
  "fmt"
)

type Hand int8

const (
	Right Hand = iota
	Left
)

type Turn int8
const (
	Player1 Turn = 1
	Player2 = 2
)

func incrementTurn(t Turn) Turn {
	if t == 1 {
		return 2
	} 
	return 1
}
	
type player struct {
    lh int
    rh  int
}

type gameState struct {
	player1 player
	player2 player
	turn Turn
}

func (gs gameState) print() {
	var player1Dec string = "  "
	var player2Dec string = "  "
	if gs.turn == Player1 {
		player1Dec = "=>"
	} else {
		player2Dec = "=>"
	}
	fmt.Println("==================================")
	fmt.Printf("==         %sPlayer 2           ==\n", player2Dec)
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", gs.player2.lh, gs.player2.rh)
	fmt.Println("==------------------------------==")
	fmt.Printf("==         %sPlayer 1           ==\n", player1Dec)
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", gs.player1.lh, gs.player1.rh)
	fmt.Println("==================================")
}

func (gs gameState) playTurn(t Turn, playerHand Hand, receiverHand Hand) gameState {
	return gs
}

func initGame() gameState {
	return gameState{
		player{1, 1},	
		player{1, 1},	
		Player1,	
	}	
}

