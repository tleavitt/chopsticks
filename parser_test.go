package main

import (
  "fmt"
  "testing"
)

func TestUnmarshalMove(t *testing.T) {
  fmt.Println("starting TestStateCopy")
  j := `{"gs":{"p1":{"lh":1,"rh":1},"p2":{"lh":1,"rh":1},"turn":"p1"},"move":{"playerHand":"rh","receiverHand":"lh"}}`
  _, _, err := parseUiMove([]byte(j))
  if err != nil {
    t.Fatal(err.Error())
  }
}