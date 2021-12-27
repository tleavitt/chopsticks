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

func (p *player) isEliminated() int {
	return p.lh == 0 && p.rh == 0
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
	// TODO: make these pointers?
	player1 player
	player2 player
	turn Turn // Turn indicates who the player is vs the receiver
}

func (gs *gameState) getPlayer() *player {
	if gs.turn == Player1 {
		return &gs.player1
	} else {
		return &gs.player2		
	}
}

func (gs *gameState) getReceiver() *player {
	if gs.turn == Player1 {
		return &gs.player2
	} else {
		return &gs.player1		
	}
}


// Update the turn variable and swap the players
func (gs *gameState) incrementTurn() *gameState {
	if gs.turn == Player1 {
		 gs.turn = Player2
	} else {
		gs.turn = Player1		
	}
	return gs
}

func (gs *gameState) toString() string {
	fmt.Sprintf("%+v\n", gs)
}

func (gs *gameState) print() {
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

// Makes a new gameastate
func copyAndPlayTurn(gs *gameState, playerHand Hand, receiverHand Hand) (*gameState, error) {
	gsCopy := *gs
	gsCopy.playTurn()
	return &gsCopy
}

// Note: mutates state
func (gs *gameState) playTurn(playerHand Hand, receiverHand Hand) (*gameState, error) {
	playerVal := gs.getPlayer().getHand(playerHand)
	if (playerVal == 0) {
		return gs, errors.New("illegalMove: attempted to play an eliminated hand")
	}
	receiverVal := gs.getReceiver().getHand(receiverHand)
	if (receiverVal == 0) {
		return gs, errors.New("illegalMove: attempted to receive on an eliminated hand")
	}
	updatedReceiverVal := (receiverVal + playerVal) % NUM_FINGERS
	gs.getReceiver().setHand(receiverHand, updatedReceiverVal)
	gs.incrementTurn()

	if DEBUG {
		gs.print()
	}
	return gs, nil
}

func initGame() gameState {
	return gameState{
		player{1, 1},	
		player{1, 1},	
		Player1,	
	}	
}

