package main
// Game play state for maintaining synchronized normalized and de-normalized states
import (
	"errors"
	"fmt"
)

type gamePlayState struct {
	state *GameState
	normalizedState *GameState
}

func (gps *gamePlayState) deepCopy() *gamePlayState {
	stateCopy := *gps.state
	normalizedStateCopy := *gps.normalizedState
	return &gamePlayState{&stateCopy, &normalizedStateCopy}
}

func (gps *gamePlayState) toString() string {
	return fmt.Sprintf("{state: %+v normalizedState: %+v", gps.state, gps.normalizedState)
}

func (gps *gamePlayState) validate() error{
	if !gps.state.copyAndNormalize().equals(gps.normalizedState) {
		return errors.New(fmt.Sprintf("State and normalized state do not match: %+v, %+v", *gps.state, *gps.normalizedState))
	} else {
		return nil
	}
}

func createGamePlayState(state *GameState) *gamePlayState {
	return &gamePlayState{
		state, state.copyAndNormalize(),
	}
}

func validateGameAndNormPlayers(gamePlayer player, normalizedPlayer player) error {
			// Extra safety belts
	if !normalizedPlayer.isNormalized() {
		return errors.New("normalizedPlayer is not normalized: " + fmt.Sprintf("%+v", normalizedPlayer))
	} 
	if gamePlayer.isNormalized() && !gamePlayer.equals(&normalizedPlayer) {
		return errors.New(fmt.Sprintf("gamePlayer (normalized) and normalizedPlayer are not synchronized: %+v, %+v", gamePlayer, normalizedPlayer))
	}
	if !gamePlayer.isNormalized() && (gamePlayer.Lh != normalizedPlayer.Rh || gamePlayer.Rh != normalizedPlayer.Lh) {
		return errors.New(fmt.Sprintf("gamePlayer and normalizedPlayer are not synchronized: %+v, %+v", gamePlayer, normalizedPlayer))
	}
	return nil
}

func getNormalizedHandForGameMoveAndPlayers(moveHand Hand, gamePlayer player, normalizedPlayer player) (Hand, error) {
	if DEBUG {
		if err := validateGameAndNormPlayers(gamePlayer, normalizedPlayer); err != nil {
			return moveHand, err
		}
	}

	// Two steps to denormalizing the move:
	// If both player's hands are equal return Left -- Left and right are equivalent in this case. Normalized moves always
	// use left and game moves can use either.
	if gamePlayer.Lh == gamePlayer.Rh {
		return Left, nil
	} else {
		// if the game player and normalized player are not the same, swap the hand. We know that the hand had to get swapped to get here.
		if gamePlayer != normalizedPlayer {
			return moveHand.invert(), nil
		} else {
			// Otherwise return the same hand
			return moveHand, nil
		}
	}
}


func getGameHandForNormalizedMoveAndPlayers(moveHand Hand, gamePlayer player, normalizedPlayer player) (Hand, error) {
	// Wackiness: the same function actually works for both cases - this is because there are only two move hands :)
	// Normalization is it's own inverse.
	return getNormalizedHandForGameMoveAndPlayers(moveHand, gamePlayer, normalizedPlayer)
}


func (gps *gamePlayState) getNormalizedMoveForGameMove(gameMove move) (move, error) {
	normalizedPlayerHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.PlayerHand, *gps.state.getPlayer(), *gps.normalizedState.getPlayer())
	if err != nil { return gameMove, err }
	normalizedReceiverHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.receiverHand, *gps.state.getReceiver(), *gps.normalizedState.getReceiver())
	if err != nil { return gameMove, err }
	return move{normalizedPlayerHand, normalizedReceiverHand}, nil
}

// TODO: DRY?
func (gps *gamePlayState) getGameMoveForNormalizedMove(normalizedMove move) (move, error) {
	gamePlayerHand, err := getGameHandForNormalizedMoveAndPlayers(normalizedMove.PlayerHand, *gps.state.getPlayer(), *gps.normalizedState.getPlayer())
	if err != nil { return normalizedMove, err }
	gameReceiverHand, err := getGameHandForNormalizedMoveAndPlayers(normalizedMove.receiverHand, *gps.state.getReceiver(), *gps.normalizedState.getReceiver())
	if err != nil { return normalizedMove, err }
	return move{gamePlayerHand, gameReceiverHand}, nil
}

func (gps *gamePlayState) applyMovesAndValidate(gameMove move, normalizedMove move) error {
	// Apply the game move to the game state
	gps.state.playMove(gameMove)
	// Apply the normalized move to the normalized state, and then normalize
	gps.normalizedState.playMove(normalizedMove)
	gps.normalizedState.normalize()
	// Validate (always, I guess)
	if err := gps.validate(); err != nil {
		return err
	}
	return nil
}

// Play a "game turn" on the game state, and synchronize the normalized state.
// Return the normalized move that we applied to the normalized state (for lookups)
func (gps *gamePlayState) playGameTurn(gameMove move) (move, error) {
	if err := gps.validate(); err != nil {
		return gameMove, err
	}
	// First determine the normalized move
	normalizedMove, err := gps.getNormalizedMoveForGameMove(gameMove)
	if err != nil { return gameMove, err }
	if err := gps.applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return gameMove, err
	}
	return normalizedMove, nil
}

// Play a "normalized turn" on the normalized state, and synchronize the game state.
// Note that the normalized move operators only on the normalized state, and the corresponding "game move"
// operates on the game state. Furthermore, the normalized state is then normalized after the move is applied.
// Return the game move that we applied to the game state (to dispaly in the UI)
func (gps *gamePlayState) playNormalizedTurn(normalizedMove move) (move, error) {
	if err := gps.validate(); err != nil {
		return normalizedMove, err
	}
	// First determine the game move
	gameMove, err := gps.getNormalizedMoveForGameMove(normalizedMove)
	if err != nil { return normalizedMove, err }
	if err := gps.applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return normalizedMove, err
	}
	return gameMove, nil
}
