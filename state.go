package main

import (
  "fmt"
  "github.com/pkg/errors"
)

const NUM_FINGERS int = 5

type Hand int8

const (
	Left Hand = iota
	Right
)

type Turn int8
const (
	Player1 Turn = 1
	Player2 = 2
)

type player struct {
    lh int
    rh  int
}

func (p *player) getHand(h Hand) int {
	if h == Left {
		return p.lh
	} else {
		return p.rh
	}
}

func (p *player) setHand(h Hand, value int) *player {
	if h == Left {
		p.lh = value
	} else {
		p.rh = value
	}
	return p
}

type gameState struct {
	player player
	receiver player
	turn Turn // Turn indicates who the player is vs the receiver
}

// Update the turn variable and swap the players
func (gs *gameState) incrementTurn() *gameState {
	if gs.turn == Player1 {
		 gs.turn = Player2
	} else {
		gs.turn = Player1		
	}
	temp := gs.player
	gs.player = gs.receiver
	gs.receiver = temp
	return gs
}

func (gs *gameState) print() {
	var player1 player
	var player2 player
	var player1Dec string = "  "
	var player2Dec string = "  "
	if gs.turn == Player1 {
		player1Dec = "=>"
		player1 = gs.player
		player2 = gs.receiver
	} else {
		player2Dec = "=>"
		player1 = gs.receiver
		player2 = gs.player
	}
	fmt.Println("==================================")
	fmt.Printf("==         %sPlayer 2           ==\n", player2Dec)
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", player2.lh, player2.rh)
	fmt.Println("==------------------------------==")
	fmt.Printf("==         %sPlayer 1           ==\n", player1Dec)
	fmt.Printf( "==      LH:%d         RH:%d       ==\n", player1.lh, player1.rh)
	fmt.Println("==================================")
}

// Note: mutates state
func (gs *gameState) playTurn(playerHand Hand, receiverHand Hand) (*gameState, error) {
	playerVal := gs.player.getHand(playerHand)
	fmt.Println(playerVal)
	if (playerVal == 0) {
		return gs, errors.New("illegalMove: attempted to play an eliminated hand")
	}
	receiverVal := gs.receiver.getHand(receiverHand)
	if (receiverVal == 0) {
		return gs, errors.New("illegalMove: attempted to receive on an eliminated hand")
	}
	updatedReceiverVal := (receiverVal + playerVal) % NUM_FINGERS
	fmt.Println(updatedReceiverVal)
	gs.receiver.setHand(receiverHand, updatedReceiverVal)
	fmt.Println(gs.receiver.getHand(receiverHand))
	gs.incrementTurn()

	gs.print()
	return gs, nil
}

func initGame() gameState {
	return gameState{
		player{1, 1},	
		player{1, 1},	
		Player1,	
	}	
}

