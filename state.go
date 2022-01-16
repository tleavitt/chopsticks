package main

import (
  "fmt"
  "errors"
  "strings"
)

type GameState struct {
	// NOTE: don't make these pointers, otherwise copying and keying doesn't work.
	Player1 Player
	Player2 Player
	T Turn // Turn indicates who the Player is vs the receiver
}

func (gs *GameState) equals(other *GameState) bool {
	return gs.Player1 == other.Player1 && gs.Player2 == other.Player2 && gs.T == other.T
}

// Maintain that the Player hands are in sorted order (smallest hand first)
func (gs *GameState) normalize() *GameState {
  gs.Player1.normalize() 
  gs.Player2.normalize() 
  return gs
}

func (gs GameState) copyAndNormalize() *GameState {
	return gs.normalize()
}

func (gs *GameState) isNormalized() bool {
	return gs.Player1.Lh <= gs.Player1.Rh && gs.Player2.Lh <= gs.Player2.Rh
}


func (gs GameState) copyAndDenormalize(swapPlayer1 bool, swapPlayer2 bool) *GameState {
	if swapPlayer1 {
		gs.Player1.Rh, gs.Player1.Lh = gs.Player1.Lh, gs.Player1.Rh
	}
	if swapPlayer2 {
		gs.Player2.Rh, gs.Player2.Lh = gs.Player2.Lh, gs.Player2.Rh
	}
	return &gs
}

func (gs *GameState) getPlayer() *Player {
	if gs.T == Player1 {
		return &gs.Player1
	} else {
		return &gs.Player2		
	}
}

func (gs *GameState) getReceiver() *Player {
	if gs.T == Player1 {
		return &gs.Player2
	} else {
		return &gs.Player1		
	}
}

// Update the turn variable and swap the Players
func (gs *GameState) incrementTurn() *GameState {
	if gs.T == Player1 {
		 gs.T = Player2
	} else {
		gs.T = Player1		
	}
	return gs
}

func (gs *GameState) toString() string {
	return fmt.Sprintf("%+v", gs)
}

func (gs *GameState) prettyString() string {
	var player1Dec string = "  "
	var player2Dec string = "  "
	if gs.T == Player1 {
		player1Dec = "=>"
	} else {
		player2Dec = "=>"
	}
	var sb strings.Builder

	sb.WriteString("==================================\n")
	sb.WriteString(fmt.Sprintf("==         %sPlayer 1           ==\n", player1Dec))
	sb.WriteString(fmt.Sprintf( "==      LH:%d         RH:%d       ==\n", gs.Player1.Lh, gs.Player1.Rh))
	sb.WriteString("==------------------------------==\n")
	sb.WriteString(fmt.Sprintf("==         %sPlayer 2           ==\n", player2Dec))
	sb.WriteString(fmt.Sprintf( "==      LH:%d         RH:%d       ==\n", gs.Player2.Lh, gs.Player2.Rh))
	sb.WriteString("==================================\n")

	return sb.String()
}

func (gs *GameState) prettyPrint() {
	fmt.Printf(gs.prettyString())
}

// Makes a new gameastate
func (gs *GameState) copyAndPlayTurn(playerHand Hand, receiverHand Hand) (*GameState, error) {
	gsCopy := *gs
	result, err := gsCopy.playTurn(playerHand, receiverHand)
	return result, err
}

// Note: mutates state
func (gs *GameState) playTurn(playerHand Hand, receiverHand Hand) (*GameState, error) {
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
func (gs *GameState) playMove(m Move) (*GameState, error) {
	return gs.playTurn(m.PlayerHand, m.ReceiverHand)
}

func (gs *GameState) isMoveValid(m Move) bool {
	return gs.getPlayer().getHand(m.PlayerHand) != 0 && gs.getReceiver().getHand(m.ReceiverHand) != 0
}

func initGame() *GameState {
	return &GameState{
		Player{1, 1},	
		Player{1, 1},	
		Player1,	
	}	
}

