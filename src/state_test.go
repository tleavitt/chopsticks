package main

import (
  "fmt"
  "testing"
)

func TestStateCopy(t *testing.T) {
  fmt.Println("starting TestStateCopy")
  gs := GameState{
    Player{2, 1}, Player{2, 1}, Player1,
  }
  gsCopy1 := gs
  gsCopy1.Player1.Lh = 1
  if gsCopy1.equals(&gs) {
    t.Fatalf("States are equal when they should differ: %+v, %+v", gsCopy1, gs)
  }

  gsCopy2 := gs
  gsCopy2.normalize()
  if gsCopy2.equals(&gs) {
    t.Fatalf("States are equal when they should differ: %+v, %+v", gsCopy2, gs)
  }

  gsCopy3 := gs
  gsNormalized := gsCopy3.copyAndNormalize()
  if !gsCopy3.equals(&gs) {
    t.Fatalf("States differ when they should be equal: %+v, %+v", gsCopy3, gs)
  }
  if gsNormalized.equals(&gs) {
    t.Fatalf("States are equal when they should differ: %+v, %+v", gsNormalized, gs)
  }
}