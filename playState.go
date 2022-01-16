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

func validateGameAndNormPlayers(gamePlayer Player, normalizedPlayer Player) error {
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

func getNormalizedHandForGameMoveAndPlayers(MoveHand Hand, gamePlayer Player, normalizedPlayer Player) (Hand, error) {
	if DEBUG {
		if err := validateGameAndNormPlayers(gamePlayer, normalizedPlayer); err != nil {
			return MoveHand, err
		}
	}

	// Two steps to denormalizing the Move:
	// If both Player's hands are equal return Left -- Left and right are equivalent in this case. Normalized Moves always
	// use left and game Moves can use either.
	if gamePlayer.Lh == gamePlayer.Rh {
		return Left, nil
	} else {
		// if the game Player and normalized Player are not the same, swap the hand. We know that the hand had to get swapped to get here.
		if gamePlayer != normalizedPlayer {
			return MoveHand.invert(), nil
		} else {
			// Otherwise return the same hand
			return MoveHand, nil
		}
	}
}


func getGameHandForNormalizedMoveAndPlayers(MoveHand Hand, gamePlayer Player, normalizedPlayer Player) (Hand, error) {
	// Wackiness: the same function actually works for both cases - this is because there are only two Move hands :)
	// Normalization is it's own inverse.
	return getNormalizedHandForGameMoveAndPlayers(MoveHand, gamePlayer, normalizedPlayer)
}


func (gps *gamePlayState) getNormalizedMoveForGameMove(gameMove Move) (Move, error) {
	normalizedPlayerHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.PlayerHand, *gps.state.getPlayer(), *gps.normalizedState.getPlayer())
	if err != nil { return gameMove, err }
	normalizedReceiverHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.ReceiverHand, *gps.state.getReceiver(), *gps.normalizedState.getReceiver())
	if err != nil { return gameMove, err }
	return Move{normalizedPlayerHand, normalizedReceiverHand}, nil
}

// TODO: DRY?
func (gps *gamePlayState) getGameMoveForNormalizedMove(normalizedMove Move) (Move, error) {
	gamePlayerHand, err := getGameHandForNormalizedMoveAndPlayers(normalizedMove.PlayerHand, *gps.state.getPlayer(), *gps.normalizedState.getPlayer())
	if err != nil { return normalizedMove, err }
	gameReceiverHand, err := getGameHandForNormalizedMoveAndPlayers(normalizedMove.ReceiverHand, *gps.state.getReceiver(), *gps.normalizedState.getReceiver())
	if err != nil { return normalizedMove, err }
	return Move{gamePlayerHand, gameReceiverHand}, nil
}

func (gps *gamePlayState) applyMovesAndValidate(gameMove Move, normalizedMove Move) error {
	// Apply the game Move to the game state
	gps.state.playMove(gameMove)
	// Apply the normalized Move to the normalized state, and then normalize
	gps.normalizedState.playMove(normalizedMove)
	gps.normalizedState.normalize()
	// Validate (always, I guess)
	if err := gps.validate(); err != nil {
		return err
	}
	return nil
}

// Play a "game turn" on the game state, and synchronize the normalized state.
// Return the normalized Move that we applied to the normalized state (for lookups)
func (gps *gamePlayState) playGameTurn(gameMove Move) (Move, error) {
	if err := gps.validate(); err != nil {
		return gameMove, err
	}
	// First determine the normalized Move
	normalizedMove, err := gps.getNormalizedMoveForGameMove(gameMove)
	if err != nil { return gameMove, err }
	if err := gps.applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return gameMove, err
	}
	return normalizedMove, nil
}

// Play a "normalized turn" on the normalized state, and synchronize the game state.
// Note that the normalized Move operators only on the normalized state, and the corresponding "game Move"
// operates on the game state. Furthermore, the normalized state is then normalized after the Move is applied.
// Return the game Move that we applied to the game state (to dispaly in the UI)
func (gps *gamePlayState) playNormalizedTurn(normalizedMove Move) (Move, error) {
	if err := gps.validate(); err != nil {
		return normalizedMove, err
	}
	// First determine the game Move
	gameMove, err := gps.getNormalizedMoveForGameMove(normalizedMove)
	if err != nil { return normalizedMove, err }
	if err := gps.applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return normalizedMove, err
	}
	return gameMove, nil
}
