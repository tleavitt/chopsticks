package main

import (
  "fmt"
  "errors"
  "strings"
)

type gameState struct {
	// TODO: make these pointers?
	player1 player
	player2 player
	turn Turn // Turn indicates who the player is vs the receiver
}

func (gs *gameState) equals(other *gameState) bool {
	return gs.player1 == other.player1 && gs.player2 == other.player2 && gs.turn == other.turn
}

// Maintain that the player hands are in sorted order (smallest hand first)
// Allows deduplication. Returns whether we swapped either of the player's hands when normalizing, 
// to enable the inverse transform.
func (gs *gameState) normalize() (*gameState, bool, bool) {
  _, swappedPlayer1 := gs.player1.normalize() 
  _, swappedPlayer2 := gs.player2.normalize() 
  return gs, swappedPlayer1, swappedPlayer2
}

func (gs *gameState) isNormalized() bool {
	return gs.player1.lh <= gs.player1.rh && gs.player2.lh <= gs.player2.rh
}

func (gs gameState) copyAndNormalize() (*gameState, bool, bool) {
	return gs.normalize()
}

func (gs gameState) copyAndDenormalize(swapPlayer1 bool, swapPlayer2 bool) *gameState {
	if swapPlayer1 {
		gs.player1.rh, gs.player1.lh = gs.player1.lh, gs.player1.rh
	}
	if swapPlayer2 {
		gs.player2.rh, gs.player2.lh = gs.player2.lh, gs.player2.rh
	}
	return &gs
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
	return fmt.Sprintf("%+v", gs)
}

func (gs *gameState) prettyString() string {
	var player1Dec string = "  "
	var player2Dec string = "  "
	if gs.turn == Player1 {
		player1Dec = "=>"
	} else {
		player2Dec = "=>"
	}
	var sb strings.Builder

	sb.WriteString("==================================\n")
	sb.WriteString(fmt.Sprintf("==         %sPlayer 1           ==\n", player1Dec))
	sb.WriteString(fmt.Sprintf( "==      LH:%d         RH:%d       ==\n", gs.player1.lh, gs.player1.rh))
	sb.WriteString("==------------------------------==\n")
	sb.WriteString(fmt.Sprintf("==         %sPlayer 2           ==\n", player2Dec))
	sb.WriteString(fmt.Sprintf( "==      LH:%d         RH:%d       ==\n", gs.player2.lh, gs.player2.rh))
	sb.WriteString("==================================\n")

	return sb.String()
}

func (gs *gameState) prettyPrint() {
	fmt.Printf(gs.prettyString())
}

// Makes a new gameastate
func copyAndPlayTurn(gs *gameState, playerHand Hand, receiverHand Hand) (*gameState, error) {
	gsCopy := *gs
	result, err := gsCopy.playTurn(playerHand, receiverHand)
	return result, err
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

	// if DEBUG {
	// 	gs.prettyPrint()
	// }
	return gs, nil
}

func initGame() gameState {
	return gameState{
		player{1, 1},	
		player{1, 1},	
		Player1,	
	}	
}

