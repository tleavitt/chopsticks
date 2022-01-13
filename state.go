package main

import (
  "fmt"
  "errors"
  "strings"
)

type gameState struct {
	// NOTE: don't make these pointers, otherwise copying and keying doesn't work.
	player1 player
	player2 player
turn Turn // Turn indicates who the player is vs the receiver
}

func (gs *gameState) equals(other *gameState) bool {
	return gs.player1 == other.player1 && gs.player2 == other.player2 && gs.turn == other.turn
}

// Maintain that the player hands are in sorted order (smallest hand first)
func (gs *gameState) normalize() *gameState {
  gs.player1.normalize() 
  gs.player2.normalize() 
  return gs
}

func (gs gameState) copyAndNormalize() *gameState {
	return gs.normalize()
}

func (gs *gameState) isNormalized() bool {
	return gs.player1.lh <= gs.player1.rh && gs.player2.lh <= gs.player2.rh
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
func (gs *gameState) copyAndPlayTurn(playerHand Hand, receiverHand Hand) (*gameState, error) {
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
	// Chopsticks update:
	updatedReceiverVal := (receiverVal + playerVal) % NUM_FINGERS
	gs.getReceiver().setHand(receiverHand, updatedReceiverVal)
	gs.incrementTurn()

	// if DEBUG {
	// 	gs.prettyPrint()
	// }
	return gs, nil
}

// Note: mutates state
func (gs *gameState) playMove(m move) (*gameState, error) {
	return gs.playTurn(m.playHand, m.receiveHand)
}

func (gs *gameState) isMoveValid(m move) bool {
	return gs.getPlayer().getHand(m.playHand) != 0 && gs.getReceiver().getHand(m.receiveHand) != 0
}

func initGame() *gameState {
	return &gameState{
		player{1, 1},	
		player{1, 1},	
		Player1,	
	}	
}

