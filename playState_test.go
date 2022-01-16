package main

import (
  "fmt"
  "testing"
)

func TestGamePlayStateDeepCopy(t *testing.T) {
  fmt.Println("starting TestGamePlayStateDeepCopy")
  gs := gameState{
    player{2, 1}, player{1, 1}, Player1,
  }
  gps := createGamePlayState(&gs)
  gps2 := gps.deepCopy()
  gs.Player1.Rh = 42
  if gps.state.Player1.Rh != 42 {
    t.Fatalf("Change to player1 not persisted: game state: %+v, game play state %+v", gs, gps.state)
  }
  if gps2.state.Player1.Rh != 1 {
    t.Fatalf("Change to player1 observed in copy: game state: %+v, game play state %+v", gs, gps2.state)
  }
}

// Damn this shit is tricky
func TestGamePlayState1(t *testing.T) {
  fmt.Println("starting TestGamePlayState1")
  prevNumFingers := setNumFingers(3)
  gs := gameState{
    player{2, 1}, player{1, 1}, Player1,
  }

  gps := createGamePlayState(&gs)

  if _, err := gps.playGameTurn(move{Left, Right}); err != nil { // state: {{2, 1},{1, 0}}
    t.Fatal(err.Error())
  } 
  if _, err := gps.playGameTurn(move{Left, Left}); err != nil { // state: {{0, 1},{1, 0}}
    t.Fatal(err.Error())
  }

  expectedGameState := &gameState{player{0, 1}, player{1, 0}, Player1,}
  expectedNormalizedState := &gameState{player{0, 1}, player{0, 1}, Player1,}

  if !gps.state.equals(expectedGameState) {
    t.Fatalf("Expected game state does match observed game state: %+v, %+v", expectedGameState, gps.state)
  }

  if !gps.normalizedState.equals(expectedNormalizedState) {
    t.Fatalf("Expected normalized state does match observed normalized state: %+v, %+v", expectedNormalizedState, gps.normalizedState)
  }
  setNumFingers(prevNumFingers)
}

func TestGamePlayState2(t *testing.T) {
  fmt.Println("starting TestGamePlayState2")
  prevNumFingers := setNumFingers(3)
  gs := gameState{
    player{1, 1}, player{1, 1}, Player1,
  }

  gps := createGamePlayState(&gs)

  if _, err := gps.playGameTurn(move{Right, Right}); err != nil { // state: {{1, 1},{1, 2}}
    t.Fatal(err.Error())
  } 
  if _, err := gps.playGameTurn(move{Left, Left}); err != nil { // state: {{2, 1},{1, 2}}
    t.Fatal(err.Error())
  }
  if _, err := gps.playGameTurn(move{Left, Right}); err != nil { // state: {{2, 1},{1, 1}}
    t.Fatal(err.Error())
  }
  if _, err := gps.playGameTurn(move{Left, Right}); err != nil { // state: {{2, 2},{1, 1}}
    t.Fatal(err.Error())
  }

  expectedGameState := &gameState{player{2, 2}, player{1, 1}, Player1,}
  expectedNormalizedState := &gameState{player{2, 2}, player{1, 1}, Player1,}

  if !gps.state.equals(expectedGameState) {
    t.Fatalf("Expected game state does match observed game state: %+v, %+v", expectedGameState, gps.state)
  }

  if !gps.normalizedState.equals(expectedNormalizedState) {
    t.Fatalf("Expected normalized state does match observed normalized state: %+v, %+v", expectedNormalizedState, gps.normalizedState)
  }
  setNumFingers(prevNumFingers)
}

func TestGamePlayState3(t *testing.T) {
  fmt.Println("starting TestGamePlayState3")
  prevNumFingers := setNumFingers(3)

  gs := gameState{
    player{2, 1}, player{2, 1}, Player1,
  }

  gps := createGamePlayState(&gs) // game state: {{2, 1},{2, 1}}, norm state: {{1, 2},{1, 2}}

  if _, err := gps.playNormalizedTurn(move{Left, Left}); err != nil { // game state: {{2, 1},{2, 2}}, norm state: {{1, 2},{2, 2}}
    t.Fatal(err.Error())
  } 
  if _, err := gps.playNormalizedTurn(move{Left, Right}); err != nil { // game state: {{1, 1},{2, 2}}, norm state: {{1, 1},{2, 2}}
    t.Fatal(err.Error())
  }
  if _, err := gps.playNormalizedTurn(move{Left, Left}); err != nil { // game state: {{1, 1},{0, 2}}, norm state: {{1, 1},{0, 2}}
    t.Fatal(err.Error())
  }


  expectedGameState := &gameState{player{1, 1}, player{0, 2}, Player2,}
  expectedNormalizedState := &gameState{player{1, 1}, player{0, 2}, Player2,}

  if !gps.state.equals(expectedGameState) {
    t.Fatalf("Expected game state does match observed game state: %+v, %+v", expectedGameState, gps.state)
  }

  if !gps.normalizedState.equals(expectedNormalizedState) {
    t.Fatalf("Expected normalized state does match observed normalized state: %+v, %+v", expectedNormalizedState, gps.normalizedState)
  }
  setNumFingers(prevNumFingers)
}

func TestGamePlayState4(t *testing.T) {
  fmt.Println("starting TestGamePlayState4")
  prevNumFingers := setNumFingers(3)


  gs := gameState{
    player{2, 1}, player{1, 2}, Player1,
  }

  gps := createGamePlayState(&gs) // game state: {{2, 1},{1, 2}}, norm state: {{1, 2},{1, 2}}

  if _, err := gps.playGameTurn(move{Left, Left}); err != nil { // game state: {{2, 1},{0, 2}}, norm state: {{1, 2},{0, 2}}
    t.Fatal(err.Error())
  } 
  if _, err := gps.playNormalizedTurn(move{Right, Left}); err != nil { // game state: {{2, 0},{0, 2}}, norm state: {{0, 2},{0, 2}}
    t.Fatal(err.Error())
  }
  if _, err := gps.playGameTurn(move{Left, Right}); err != nil { // game state: {{2, 0},{0, 1}}, norm state: {{0, 2},{0, 1}}
    t.Fatal(err.Error())
  }


  expectedGameState := &gameState{player{2, 0}, player{0, 1}, Player2,}
  expectedNormalizedState := &gameState{player{0, 2}, player{0, 1}, Player2,}

  if !gps.state.equals(expectedGameState) {
    t.Fatalf("Expected game state does match observed game state: %+v, %+v", expectedGameState, gps.state)
  }

  if !gps.normalizedState.equals(expectedNormalizedState) {
    t.Fatalf("Expected normalized state does match observed normalized state: %+v, %+v", expectedNormalizedState, gps.normalizedState)
  }
  setNumFingers(prevNumFingers)
}