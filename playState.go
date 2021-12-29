package main
// Game play state for maintaining synchronized normalized and de-normalized states

type gamePlayState struct {
	state *gameState
	normalizedState *gameState
}

func (gps *gamePlayState) validate() error{
	if !gps.state.copyAndNormalize().equals(gps.normalizedState) {
		return errors.New("State and normalized state do not match: %+v, %+v", *gps.state, *gps.normalizedState)
	} else {
		return nil
	}
}

func createGamePlayState(state *gameState) {
	return gamePlayState{
		state, state.copyAndNormalize()
	}
}

func validateGameAndNormPlayers(gamePlayer player, normalizedPlayer player) error {
			// Extra safety belts
	if !normalizedPlayer.isNormalized() {
		return errors.New("normalizedPlayer is not normalized: " + fmt.Printf("%+v", normalizedPlayer))
	} 
	if gamePlayer.isNormalized() && gamePlayer != normalizedPlayer {
		return errors.New(fmt.Printf("gamePlayer and normalizedPlayer are not synchronized: %+v, %+v", gamePlayer, normalizedPlayer))
	} else if gamePlayer.lh != normalizedPlayer.rh || gamePlayer.rh != normalizedPlayer.lh {
		return errors.New(fmt.Printf("gamePlayer and normalizedPlayer are not synchronized: %+v, %+v", gamePlayer, normalizedPlayer))
	}
	return nil
}

func getNormalizedHandForGameMoveAndPlayers(normalizedHand Hand, gamePlayer player, normalizedPlayer player) (Hand, error) {
	if DEBUG {
		if err := validateGameAndNormPlayers(gamePlayer, normalizedPlayer); err != nil {
			return moveHand, err
		}
	}

	// Two steps to denormalizing the move:
	// If both player's hands are equal return Left -- Left and right are equivalent in this case. Normalized moves always
	// use left and game moves can use either.
	if gamePlayer.lh == gamePlayer.rh {
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
	return getNormalizedHandForGameMoveAndPlayers()
}


func (gps *gamePlayState) getNormalizedMoveForGameMove(gameMove move) (move, error) {
	normalizedPlayerHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.playHand, gps.state.getPlayer(), gps.normalizedState.getPlayer())
	if err != nil { return err }
	normalizedReceiverHand, err := getNormalizedHandForGameMoveAndPlayers(gameMove.receiveHand, gps.state.getReceiver(), gps.normalizedState.getReceiver())
	if err != nil { return err }
	return move{normalizedPlayerHand, normalizedReceiverHand}, nil
}

// TODO: DRY?
func (gps *gamePlayState) getGameMoveForNormalizedMove(normalizedMove move) move {
	gamePlayerHand, err := getGameHandForNormalizedMoveAndPlayers(gameMove.playHand, gps.state.getPlayer(), gps.normalizedState.getPlayer())
	if err != nil { return err }
	gameReceiverHand, err := getGameHandForNormalizedMoveAndPlayers(gameMove.receiveHand, gps.state.getReceiver(), gps.normalizedState.getReceiver())
	if err != nil { return err }
	return move{gamePlayerHand, gameReceiverHand}, nil
}

func (gps *gamePlayState) applyMovesAndValidate(gameMove move, normalizedMove move) error {
	// Apply the game move to the game state
	gps.state.playTurn(gameMove)
	// Apply the normalized move to the normalized state
	gps.normalizedState.playTurn(normalizedMove)
	// Validate (always, I guess)
	if err := gps.validate(); err != nil {
		return err
	}
}

// Play a "game turn" on the game state, and synchronize the normalized state.
// Return the normalized move that we applied to the normalized state (for lookups)
func (gps *gamePlayState) playGameTurn(gameMove move) (move, error) {
	if err := gps.validate(); err != nil {
		return gameMove, err
	}
	// First determine the normalized move
	normalizedMove, err := gps.getNormalizedMoveForGameMove(gameMove)
	if err != nil { return err }
	if err := applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return gameMove, nil
	}
	return normalizedMove, nil
}

// Play a "normalized turn" on the normalized state, and synchronize the game state.
// Return the game move that we applied to the game state (to dispaly in the UI)
func (gps *gamePlayState) playNormalizedTurn(normalizedMove move) (move, error) {
	if err := gps.validate(); err != nil {
		return normalizedMove, err
	}
	// First determine the game move
	gameMove, err := gps.getNormalizedMoveForGameMove(gameMove)
	if err != nil { return err }
	if err := applyMovesAndValidate(gameMove, normalizedMove); err != nil {
		return normalizedMove, nil
	}
	return gameMove, nil
}
